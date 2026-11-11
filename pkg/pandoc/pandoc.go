// Package pandoc builds the Pandoc command that converts a rendered
// agreement (Pandoc-flavored Markdown) to DOCX or PDF. It does not run the
// command; the CLI does, so this package stays free of process and file
// I/O and can be tested without Pandoc installed.
package pandoc

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Target is an output format.
type Target string

// Targets.
const (
	TargetDOCX Target = "docx"
	TargetPDF  Target = "pdf"
)

// DefaultPDFEngine handles Unicode punctuation without extra setup and,
// unlike xelatex and pdflatex, writes the same PDF ID on every run, so PDF
// output is byte-for-byte reproducible.
const DefaultPDFEngine = "lualatex"

// ErrUnknownTarget is returned for an unsupported output format.
var ErrUnknownTarget = errors.New("unknown target")

// TargetFromPath returns the target implied by an output file extension.
func TargetFromPath(path string) (Target, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".docx":
		return TargetDOCX, nil
	case ".pdf":
		return TargetPDF, nil
	default:
		return "", fmt.Errorf("%w: %q (want .docx or .pdf)", ErrUnknownTarget, path)
	}
}

// Options describe one conversion.
type Options struct {
	Input  string
	Output string
	Target Target
	// PDFEngine is the LaTeX engine for PDF output; DefaultPDFEngine when
	// empty.
	PDFEngine string
	// ReferenceDoc is an optional DOCX whose styles the output uses.
	ReferenceDoc string
	// Date pins the timestamps Pandoc and LaTeX embed in the output
	// (YYYY-MM-DD), so the same input produces the same file. Use the
	// agreement's date.
	Date string
	// Extra are additional Pandoc arguments, appended last.
	Extra []string
}

// Command returns the Pandoc arguments (without the program name) and the
// extra environment variables for o.
func Command(o Options) (args, env []string, err error) {
	if o.Input == "" || o.Output == "" {
		return nil, nil, errors.New("pandoc: input and output are required")
	}
	args = []string{"--from=markdown", "--standalone", "--output=" + o.Output}
	switch o.Target {
	case TargetDOCX:
		args = append(args, "--to=docx")
		if o.ReferenceDoc != "" {
			args = append(args, "--reference-doc="+o.ReferenceDoc)
		}
	case TargetPDF:
		engine := o.PDFEngine
		if engine == "" {
			engine = DefaultPDFEngine
		}
		args = append(args, "--pdf-engine="+engine)
	default:
		return nil, nil, fmt.Errorf("%w: %q", ErrUnknownTarget, string(o.Target))
	}
	args = append(args, o.Extra...)
	args = append(args, o.Input)

	if o.Date != "" {
		t, err := time.Parse(time.DateOnly, o.Date)
		if err != nil {
			return nil, nil, fmt.Errorf("pandoc: date %q: %w", o.Date, err)
		}
		env = []string{
			"SOURCE_DATE_EPOCH=" + strconv.FormatInt(t.Unix(), 10),
			// Makes TeX engines use SOURCE_DATE_EPOCH for \today and the
			// PDF creation date as well as the PDF ID.
			"FORCE_SOURCE_DATE=1",
		}
	}
	return slices.Clip(args), env, nil
}
