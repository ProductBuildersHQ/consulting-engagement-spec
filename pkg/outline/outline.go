// Package outline defines the blocks an agreement is made of, the
// references an outline uses to name them, and the built-in outlines.
//
// A block is a part of an agreement the renderer can print, such as
// "objectives" or "scope.excluded". A reference names a block, optionally
// narrowed by a filter: a party role ("responsibilities.client") or a
// milestone kind ("milestones.closeout").
package outline

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
)

// FilterKind is what a block can be narrowed by.
type FilterKind int

// Filter kinds.
const (
	// NoFilter means the block takes no filter.
	NoFilter FilterKind = iota
	// RoleFilter narrows by party role, such as "client".
	RoleFilter
	// MilestoneKindFilter narrows by milestone kind, such as "kickoff".
	MilestoneKindFilter
)

// Block is a part of an agreement the renderer can print.
type Block struct {
	// Name is the block's reference, such as "scope.included".
	Name string
	// Title is the block's heading when it is not a section's first block.
	Title string
	// Parent is the block that contains this one, such as "scope" for
	// "scope.included". Empty for a top-level block.
	Parent string
	Filter FilterKind
}

var blocks = []Block{
	{Name: "summary", Title: "Summary"},
	{Name: "background", Title: "Background"},
	{Name: "objectives", Title: "Objectives"},
	{Name: "hypotheses", Title: "Hypotheses"},
	{Name: "scope", Title: "Scope"},
	{Name: "scope.included", Title: "In scope", Parent: "scope"},
	{Name: "scope.excluded", Title: "Out of scope", Parent: "scope"},
	{Name: "approach", Title: "Approach"},
	{Name: "phases", Title: "Phases"},
	{Name: "deliverables", Title: "Deliverables"},
	{Name: "deliverables.list", Title: "Deliverables", Parent: "deliverables"},
	{Name: "milestones", Title: "Milestones", Filter: MilestoneKindFilter},
	{Name: "responsibilities", Title: "Responsibilities", Filter: RoleFilter},
	{Name: "people", Title: "Team and contacts"},
	{Name: "governance", Title: "Working together"},
	{Name: "governance.contacts", Title: "Points of contact", Filter: RoleFilter},
	{Name: "assumptions", Title: "Assumptions"},
	{Name: "risks", Title: "Risks"},
	{Name: "successCriteria", Title: "Measures of success"},
	{Name: "acceptance", Title: "Review and acceptance"},
	{Name: "closeout", Title: "Closeout"},
	{Name: "commercials", Title: "Fees and expenses"},
	{Name: "changeControl", Title: "Changes to this agreement"},
	{Name: "terms", Title: "Terms"},
	{Name: "openQuestions", Title: "Open questions"},
	{Name: "signatures", Title: "Signatures", Filter: RoleFilter},
}

// Blocks returns every block, in the order of the full layout.
func Blocks() []Block { return slices.Clone(blocks) }

// LookupBlock returns the block with the given name.
func LookupBlock(name string) (Block, bool) {
	for _, b := range blocks {
		if b.Name == name {
			return b, true
		}
	}
	return Block{}, false
}

// Ref is a parsed block reference.
type Ref struct {
	Block Block
	// Value is the filter value, such as "client" or "kickoff"; empty when
	// the reference has no filter.
	Value string
}

// String returns the reference as written.
func (r Ref) String() string {
	if r.Value == "" {
		return r.Block.Name
	}
	return r.Block.Name + "." + r.Value
}

// Top returns the name of the top-level block r falls under: its block's
// parent, or the block itself.
func (r Ref) Top() string {
	if r.Block.Parent != "" {
		return r.Block.Parent
	}
	return r.Block.Name
}

// ParseRef parses a block reference such as "responsibilities.client".
func ParseRef(s string) (Ref, error) {
	if b, ok := LookupBlock(s); ok {
		return Ref{Block: b}, nil
	}
	i := strings.LastIndexByte(s, '.')
	if i < 0 {
		return Ref{}, fmt.Errorf("unknown block %q", s)
	}
	b, ok := LookupBlock(s[:i])
	if !ok {
		return Ref{}, fmt.Errorf("unknown block %q", s)
	}
	v := s[i+1:]
	switch b.Filter {
	case RoleFilter:
		if !engagement.PartyRole(v).Valid() {
			return Ref{}, fmt.Errorf("%s: unknown party role %q", s, v)
		}
	case MilestoneKindFilter:
		if !engagement.MilestoneKind(v).Valid() {
			return Ref{}, fmt.Errorf("%s: unknown milestone kind %q", s, v)
		}
	default:
		return Ref{}, fmt.Errorf("%s: block %q takes no filter", s, b.Name)
	}
	return Ref{Block: b, Value: v}, nil
}

// Overlaps reports whether two references can print the same content: the
// same block or one containing the other, with filters that do not exclude
// each other.
func Overlaps(a, b Ref) bool {
	if a.Top() != b.Top() {
		return false
	}
	if a.Block.Name != b.Block.Name && a.Block.Parent != "" && b.Block.Parent != "" {
		return false
	}
	return a.Value == "" || b.Value == "" || a.Value == b.Value
}

// HasContent reports whether r prints anything for a. Required parts of an
// agreement (scope, deliverables, responsibilities, commercials) always do,
// as the renderer always prints them.
func HasContent(a *engagement.Agreement, r Ref) bool {
	nonBlank := func(s string) bool { return strings.TrimSpace(s) != "" }
	switch r.Block.Name {
	case "summary":
		return nonBlank(a.Summary)
	case "background":
		return nonBlank(a.Background)
	case "objectives":
		return len(a.Objectives) > 0
	case "hypotheses":
		return len(a.Hypotheses) > 0
	case "scope", "commercials":
		return true
	case "scope.included":
		return len(a.Scope.Included) > 0
	case "scope.excluded":
		return len(a.Scope.Excluded) > 0
	case "approach":
		return nonBlank(a.Approach)
	case "phases":
		return len(a.Phases) > 0
	case "deliverables", "deliverables.list":
		return true
	case "milestones":
		return len(Milestones(a, r.Value)) > 0
	case "responsibilities":
		return r.Value == "" || len(Responsibilities(a, r.Value)) > 0
	case "people":
		return len(a.People) > 0
	case "governance":
		return a.Governance != nil
	case "governance.contacts":
		return len(Contacts(a, r.Value)) > 0
	case "assumptions":
		return len(a.Assumptions) > 0
	case "risks":
		return len(a.Risks) > 0
	case "successCriteria":
		return len(a.SuccessCriteria) > 0
	case "acceptance":
		return a.Acceptance != nil
	case "closeout":
		return a.Closeout != nil
	case "changeControl":
		return nonBlank(a.ChangeControl)
	case "terms":
		return len(a.Terms) > 0
	case "openQuestions":
		return len(a.OpenQuestions) > 0
	case "signatures":
		return len(Signatures(a, r.Value)) > 0
	default:
		return false
	}
}

// partyRole returns the role of the party with the given ID.
func partyRole(a *engagement.Agreement, id string) engagement.PartyRole {
	p, _ := a.Party(id) // an unknown party has no role, so no role filter matches it
	return p.Role
}

// Milestones returns the milestones of the given kind, or all when kind is
// empty.
func Milestones(a *engagement.Agreement, kind string) []engagement.Milestone {
	var out []engagement.Milestone
	for _, m := range a.Milestones {
		if kind == "" || string(m.Kind) == kind {
			out = append(out, m)
		}
	}
	return out
}

// Responsibilities returns the responsibilities of parties with the given
// role, or all when role is empty.
func Responsibilities(a *engagement.Agreement, role string) []engagement.Responsibility {
	var out []engagement.Responsibility
	for _, r := range a.Responsibilities {
		if role == "" || string(partyRole(a, r.Party)) == role {
			out = append(out, r)
		}
	}
	return out
}

// Contacts returns the points of contact of parties with the given role,
// or all when role is empty.
func Contacts(a *engagement.Agreement, role string) []engagement.Contact {
	if a.Governance == nil {
		return nil
	}
	var out []engagement.Contact
	for _, c := range a.Governance.Contacts {
		if role == "" || string(partyRole(a, c.Party)) == role {
			out = append(out, c)
		}
	}
	return out
}

// Signatures returns the signature blocks of parties with the given role,
// or all when role is empty.
func Signatures(a *engagement.Agreement, role string) []engagement.Signature {
	var out []engagement.Signature
	for _, s := range a.Signatures {
		if role == "" || string(partyRole(a, s.Party)) == role {
			out = append(out, s)
		}
	}
	return out
}
