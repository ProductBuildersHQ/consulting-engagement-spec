// Package crosswalk shows how an agreement answers another outline: for
// each section of the target outline, the sections of the rendered
// agreement that cover it, and the sections the target does not ask for.
//
// A typical use is a full agreement next to the shorter form a program
// office or a client expects: the crosswalk lets a reader who knows the
// form find each of its sections in the agreement, and shows what the
// agreement adds.
package crosswalk

import (
	"fmt"
	"strings"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/outline"
)

// Section is a section of the rendered agreement.
type Section struct {
	// Number is the section's number as rendered, or 0 when the layout is
	// not numbered.
	Number  int    `json:"number,omitempty"`
	Heading string `json:"heading"`
}

// Label returns "§3 Objectives", or the heading alone when unnumbered.
func (s Section) Label() string {
	if s.Number == 0 {
		return s.Heading
	}
	return fmt.Sprintf("§%d %s", s.Number, s.Heading)
}

// Row is one section of the target outline.
type Row struct {
	Heading string `json:"heading"`
	// Blocks are the target section's block references.
	Blocks []string `json:"blocks"`
	// Sections are the sections of the rendered agreement that answer it.
	Sections []Section `json:"sections"`
	// Answered is true when the agreement has content for the section.
	Answered bool `json:"answered"`
}

// Crosswalk relates a rendered agreement to a target outline.
type Crosswalk struct {
	// Layout is the title of the outline the agreement is rendered in.
	Layout string `json:"layout"`
	// Target is the title of the outline being answered.
	Target string `json:"target"`
	Rows   []Row  `json:"rows"`
	// Beyond lists sections of the rendered agreement that answer no
	// section of the target: what the agreement adds.
	Beyond []Section `json:"beyond"`
}

type rendered struct {
	Section
	refs []outline.Ref
}

// Build relates a, rendered in layout, to target.
func Build(a *engagement.Agreement, layout, target *engagement.Outline) (Crosswalk, error) {
	// Number the layout's sections the way the renderer does: a section
	// prints only when one of its blocks has content.
	var secs []rendered
	n := 0
	for _, s := range layout.Sections {
		refs, err := parse(layout, s)
		if err != nil {
			return Crosswalk{}, err
		}
		shown := false
		for _, r := range refs {
			if outline.HasContent(a, r) {
				shown = true
				break
			}
		}
		if !shown {
			continue
		}
		rs := rendered{Section: Section{Heading: s.Heading}, refs: refs}
		if layout.IsNumbered() {
			n++
			rs.Number = n
		}
		secs = append(secs, rs)
	}

	cw := Crosswalk{Layout: layout.Title, Target: target.Title}
	used := make([]bool, len(secs))
	for _, s := range target.Sections {
		refs, err := parse(target, s)
		if err != nil {
			return Crosswalk{}, err
		}
		row := Row{Heading: s.Heading, Blocks: s.From}
		for _, r := range refs {
			if outline.HasContent(a, r) {
				row.Answered = true
			}
		}
		for i, rs := range secs {
			if overlaps(refs, rs.refs, a) {
				row.Sections = append(row.Sections, rs.Section)
				used[i] = true
			}
		}
		if len(row.Sections) == 0 {
			row.Answered = false
		}
		cw.Rows = append(cw.Rows, row)
	}
	for i, rs := range secs {
		if !used[i] {
			cw.Beyond = append(cw.Beyond, rs.Section)
		}
	}
	return cw, nil
}

func parse(o *engagement.Outline, s engagement.OutlineSection) ([]outline.Ref, error) {
	refs := make([]outline.Ref, 0, len(s.From))
	for _, name := range s.From {
		r, err := outline.ParseRef(name)
		if err != nil {
			return nil, fmt.Errorf("crosswalk: outline %s, section %q: %w", o.ID, s.Heading, err)
		}
		refs = append(refs, r)
	}
	return refs, nil
}

// overlaps reports whether a target section and a rendered section share
// content: some pair of their references overlaps, and the rendered one
// has content.
func overlaps(target, layout []outline.Ref, a *engagement.Agreement) bool {
	for _, t := range target {
		for _, l := range layout {
			if outline.Overlaps(t, l) && outline.HasContent(a, l) {
				return true
			}
		}
	}
	return false
}

// Markdown renders the crosswalk as a table, then the sections the
// agreement adds.
func (c Crosswalk) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Each section of the %s and where this document answers it.\n\n", c.Target)
	fmt.Fprintf(&b, "| %s | In this document |\n|----------------------------|----------------------------------------|\n", cell(c.Target))
	for _, r := range c.Rows {
		where := "Not addressed"
		if r.Answered {
			labels := make([]string, len(r.Sections))
			for i, s := range r.Sections {
				labels[i] = s.Label()
			}
			where = strings.Join(labels, "; ")
		}
		fmt.Fprintf(&b, "| %s | %s |\n", cell(r.Heading), cell(where))
	}
	b.WriteString("\n")
	if len(c.Beyond) > 0 {
		fmt.Fprintf(&b, "This document also covers, beyond the %s:\n\n", c.Target)
		for _, s := range c.Beyond {
			fmt.Fprintf(&b, "- %s\n", s.Label())
		}
		b.WriteString("\n")
	}
	return b.String()
}

func cell(s string) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(s), " "), "|", `\|`)
}
