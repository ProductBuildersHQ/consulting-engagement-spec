package lint

import (
	"fmt"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/outline"
)

// Outline rule IDs.
const (
	RuleUnknownBlock     = "unknown-block"
	RuleDuplicateHeading = "duplicate-heading"
	RuleEmptySection     = "empty-section"
)

// Outline lints an outline on its own: every block reference must resolve,
// and no two sections may share a heading.
func Outline(o *engagement.Outline) Result {
	var out []Finding
	seen := map[string]int{}
	for i, s := range o.Sections {
		path := fmt.Sprintf("sections/%d", i)
		if prev, ok := seen[s.Heading]; ok {
			out = append(out, Finding{Error, RuleDuplicateHeading, path + "/heading",
				fmt.Sprintf("heading %q is already used by sections/%d", s.Heading, prev)})
		} else {
			seen[s.Heading] = i
		}
		for j, ref := range s.From {
			if _, err := outline.ParseRef(ref); err != nil {
				out = append(out, Finding{Error, RuleUnknownBlock, fmt.Sprintf("%s/from/%d", path, j), err.Error()})
			}
		}
	}
	return Result{Findings: out}
}

// Layout lints an agreement rendered in an outline: it warns for each
// section that has nothing to print for this agreement, so a reader
// expecting that section knows it will be missing.
func Layout(a *engagement.Agreement, o *engagement.Outline) Result {
	res := Outline(o)
	for i, s := range o.Sections {
		empty := true
		for _, ref := range s.From {
			r, err := outline.ParseRef(ref)
			if err == nil && outline.HasContent(a, r) {
				empty = false
				break
			}
		}
		if empty {
			res.Findings = append(res.Findings, Finding{Warning, RuleEmptySection, fmt.Sprintf("sections/%d", i),
				fmt.Sprintf("section %q has nothing to print for this agreement", s.Heading)})
		}
	}
	return res
}
