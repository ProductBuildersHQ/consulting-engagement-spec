package outline

import (
	"slices"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
)

// Names of the built-in outlines.
const (
	// NameFull is the complete agreement, the default layout.
	NameFull = "full"
	// NameSummary is a one-page plan.
	NameSummary = "summary"
)

func section(heading string, from ...string) engagement.OutlineSection {
	return engagement.OutlineSection{Heading: heading, From: from}
}

// Full returns the complete agreement layout. Rendering in it is the same
// as rendering with no outline.
func Full() *engagement.Outline {
	return &engagement.Outline{
		SchemaVersion: "0.1",
		Kind:          engagement.KindOutline,
		ID:            NameFull,
		Title:         "Full agreement",
		Sections: []engagement.OutlineSection{
			section("Summary", "summary"),
			section("Background", "background"),
			section("Objectives", "objectives"),
			section("Hypotheses", "hypotheses"),
			section("Scope", "scope"),
			section("Approach", "approach", "phases"),
			section("Deliverables", "deliverables"),
			section("Milestones", "milestones"),
			section("Responsibilities", "responsibilities"),
			section("Team and contacts", "people"),
			section("Working together", "governance", "governance.contacts"),
			section("Assumptions", "assumptions"),
			section("Risks", "risks"),
			section("Measures of success", "successCriteria"),
			section("Review and acceptance", "acceptance"),
			section("Closeout", "closeout"),
			section("Fees and expenses", "commercials"),
			section("Changes to this agreement", "changeControl"),
			section("Terms", "terms"),
			section("Open questions", "openQuestions"),
			section("Signatures", "signatures"),
		},
	}
}

// Summary returns the one-page plan layout.
func Summary() *engagement.Outline {
	return &engagement.Outline{
		SchemaVersion: "0.1",
		Kind:          engagement.KindOutline,
		ID:            NameSummary,
		Title:         "One-page plan",
		Sections: []engagement.OutlineSection{
			section("Summary", "summary"),
			section("Objectives", "objectives"),
			section("Hypotheses and assumptions", "hypotheses", "assumptions"),
			section("Scope", "scope.included"),
			section("Client responsibilities", "responsibilities.client", "governance.contacts.client"),
			section("Deliverables", "deliverables.list"),
			section("Schedule", "phases", "milestones"),
			section("Signatures", "signatures"),
		},
	}
}

// Builtin returns the built-in outline with the given name.
func Builtin(name string) (*engagement.Outline, bool) {
	switch name {
	case NameFull:
		return Full(), true
	case NameSummary:
		return Summary(), true
	default:
		return nil, false
	}
}

// BuiltinNames returns the names of the built-in outlines, sorted.
func BuiltinNames() []string {
	n := []string{NameFull, NameSummary}
	slices.Sort(n)
	return n
}
