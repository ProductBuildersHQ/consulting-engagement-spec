// Package cli implements the cespec command. Commands are thin: they parse
// flags, call the library packages, and write results.
package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagementio"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/lint"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/pandoc"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/render"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/schema"
)

// errLintFailed is returned when lint reports errors (or warnings with
// --strict), so the process exits non-zero without a second message.
var errLintFailed = errors.New("lint failed")

// NewRoot returns the root command.
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "cespec",
		Short:         "Author, check, and render consulting engagement agreements",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(validateCmd(), lintCmd(), fmtCmd(), digestCmd(), renderCmd(), exportCmd(), schemaCmd())
	root.SetErrPrefix("cespec:")
	return wrapErrors(root)
}

// wrapErrors prints command errors once, except errLintFailed, whose
// findings are already printed.
func wrapErrors(root *cobra.Command) *cobra.Command {
	for _, c := range root.Commands() {
		run := c.RunE
		if run == nil {
			continue
		}
		c.RunE = func(cmd *cobra.Command, args []string) error {
			err := run(cmd, args)
			if err != nil && !errors.Is(err, errLintFailed) {
				fmt.Fprintln(cmd.ErrOrStderr(), "cespec:", err)
			}
			return err
		}
	}
	return root
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate FILE...",
		Short: "Check agreements against the JSON Schema",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, path := range args {
				if _, err := engagementio.LoadAgreement(path); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s: valid\n", path)
			}
			return nil
		},
	}
}

func lintCmd() *cobra.Command {
	var strict, quiet bool
	c := &cobra.Command{
		Use:   "lint FILE...",
		Short: "Check references, dates, and completeness",
		Long: "Lint checks what the schema cannot: unique IDs, references that resolve, dates in order,\n" +
			"and gaps that tend to cause disputes, such as nothing out of scope or no client responsibilities.\n" +
			"It exits non-zero on errors, or on warnings with --strict.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			failed := false
			out := cmd.OutOrStdout()
			for _, path := range args {
				a, err := engagementio.LoadAgreement(path)
				if err != nil {
					return err
				}
				res := lint.Agreement(a)
				for _, f := range res.Findings {
					if quiet && f.Severity == lint.Warning {
						continue
					}
					fmt.Fprintf(out, "%s: %s\n", path, f)
				}
				fmt.Fprintf(out, "%s: %d errors, %d warnings\n", path, res.Errors(), res.Warnings())
				if res.Errors() > 0 || (strict && res.Warnings() > 0) {
					failed = true
				}
			}
			if failed {
				return errLintFailed
			}
			return nil
		},
	}
	c.Flags().BoolVar(&strict, "strict", false, "fail on warnings as well as errors")
	c.Flags().BoolVarP(&quiet, "quiet", "q", false, "print errors only")
	return c
}

func fmtCmd() *cobra.Command {
	var check bool
	c := &cobra.Command{
		Use:   "fmt FILE...",
		Short: "Rewrite agreements in canonical form",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var unformatted []string
			for _, path := range args {
				a, err := engagementio.LoadAgreement(path)
				if err != nil {
					return err
				}
				f, err := engagementio.FormatFromPath(path)
				if err != nil {
					return err
				}
				want, err := engagementio.Canonical(a, f)
				if err != nil {
					return err
				}
				got, err := os.ReadFile(path) //nolint:gosec // the user names the file to format
				if err != nil {
					return err
				}
				if string(got) == string(want) {
					continue
				}
				if check {
					unformatted = append(unformatted, path)
					continue
				}
				if err := os.WriteFile(path, want, 0o644); err != nil { //nolint:gosec // agreements are ordinary documents
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "formatted %s\n", path)
			}
			if len(unformatted) > 0 {
				return fmt.Errorf("not in canonical form: %s", strings.Join(unformatted, ", "))
			}
			return nil
		},
	}
	c.Flags().BoolVar(&check, "check", false, "report files that are not canonical instead of rewriting them")
	return c
}

func digestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "digest FILE",
		Short: "Print the content digest printed on the rendered document",
		Long: "The digest is a SHA-256 of the agreement's canonical JSON, leaving out the status and the\n" +
			"signing records. It is the same for YAML and JSON input and does not change when the agreement\n" +
			"is signed, so it ties a signed PDF to the exact source it was rendered from.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := engagementio.LoadAgreement(args[0])
			if err != nil {
				return err
			}
			d, err := engagementio.ContentDigest(a)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), d)
			return nil
		},
	}
}

func renderMarkdown(path string, opts render.Options) ([]byte, string, error) {
	a, err := engagementio.LoadAgreement(path)
	if err != nil {
		return nil, "", err
	}
	if res := lint.Agreement(a); res.Errors() > 0 {
		return nil, "", fmt.Errorf("%s: %d lint errors; run cespec lint", path, res.Errors())
	}
	md, err := render.Markdown(a, opts)
	if err != nil {
		return nil, "", err
	}
	return md, a.Date, nil
}

func renderCmd() *cobra.Command {
	var flavor, out string
	var noDigest bool
	c := &cobra.Command{
		Use:   "render FILE",
		Short: "Render an agreement as Markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			md, _, err := renderMarkdown(args[0], render.Options{Flavor: render.Flavor(flavor), OmitDigest: noDigest})
			if err != nil {
				return err
			}
			if out == "" || out == "-" {
				_, err := cmd.OutOrStdout().Write(md)
				return err
			}
			return os.WriteFile(out, md, 0o644) //nolint:gosec // rendered documents are ordinary files
		},
	}
	c.Flags().BoolVar(&noDigest, "no-digest", false, "leave the content digest out of the document")
	c.Flags().StringVar(&flavor, "flavor", string(render.FlavorPandoc), "Markdown flavor: pandoc or gfm")
	c.Flags().StringVarP(&out, "out", "o", "", "output file (default stdout)")
	return c
}

func exportCmd() *cobra.Command {
	var out, engine, refDoc, pandocBin string
	var keepMarkdown, noDigest bool
	c := &cobra.Command{
		Use:   "export FILE --out OUT.docx|OUT.pdf",
		Short: "Export an agreement to DOCX or PDF with Pandoc",
		Long: "Export renders the agreement as Pandoc Markdown and runs Pandoc. PDF output needs a LaTeX\n" +
			"engine (lualatex by default). Timestamps are pinned to the agreement's date, so the same input\n" +
			"produces the same file.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if out == "" {
				return errors.New("--out is required")
			}
			target, err := pandoc.TargetFromPath(out)
			if err != nil {
				return err
			}
			md, date, err := renderMarkdown(args[0], render.Options{Flavor: render.FlavorPandoc, OmitDigest: noDigest})
			if err != nil {
				return err
			}
			mdPath := strings.TrimSuffix(out, filepath.Ext(out)) + ".md"
			if !keepMarkdown {
				dir, err := os.MkdirTemp("", "cespec-")
				if err != nil {
					return err
				}
				defer func() {
					if err := os.RemoveAll(dir); err != nil {
						fmt.Fprintln(cmd.ErrOrStderr(), "cespec: remove temp dir:", err)
					}
				}()
				mdPath = filepath.Join(dir, "agreement.md")
			}
			if err := os.WriteFile(mdPath, md, 0o644); err != nil { //nolint:gosec // rendered documents are ordinary files
				return err
			}
			pargs, env, err := pandoc.Command(pandoc.Options{
				Input: mdPath, Output: out, Target: target,
				PDFEngine: engine, ReferenceDoc: refDoc, Date: date,
			})
			if err != nil {
				return err
			}
			run := exec.CommandContext(cmd.Context(), pandocBin, pargs...) //nolint:gosec // runs the user-chosen pandoc binary with built arguments
			run.Env = append(os.Environ(), env...)
			run.Stdout = cmd.ErrOrStderr()
			run.Stderr = cmd.ErrOrStderr()
			if err := run.Run(); err != nil {
				return fmt.Errorf("pandoc: %w", err)
			}
			if target == pandoc.TargetDOCX {
				if err := normalizeDOCX(out, date); err != nil {
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", out)
			return nil
		},
	}
	c.Flags().BoolVar(&noDigest, "no-digest", false, "leave the content digest out of the document")
	c.Flags().StringVarP(&out, "out", "o", "", "output file, .docx or .pdf (required)")
	c.Flags().StringVar(&engine, "pdf-engine", pandoc.DefaultPDFEngine, "LaTeX engine for PDF output")
	c.Flags().StringVar(&refDoc, "reference-doc", "", "DOCX whose styles the output uses")
	c.Flags().StringVar(&pandocBin, "pandoc", "pandoc", "path to the pandoc binary")
	c.Flags().BoolVar(&keepMarkdown, "keep-markdown", false, "also write the Pandoc Markdown next to the output")
	return c
}

func schemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "schema [NAME]",
		Short: "Print an embedded JSON Schema, or list them",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				for _, n := range schema.Names() {
					fmt.Fprintln(cmd.OutOrStdout(), n)
				}
				return nil
			}
			data, err := schema.Get(args[0])
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
}

// normalizeDOCX pins every entry time in the DOCX at path to the
// agreement's date, so repeated exports produce identical bytes.
func normalizeDOCX(path, date string) error {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return fmt.Errorf("agreement date %q: %w", date, err)
	}
	data, err := os.ReadFile(path) //nolint:gosec // the file pandoc just wrote
	if err != nil {
		return err
	}
	norm, err := pandoc.NormalizeZip(data, t)
	if err != nil {
		return err
	}
	return os.WriteFile(path, norm, 0o644) //nolint:gosec // rendered documents are ordinary files
}
