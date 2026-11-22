package outline

import (
	"testing"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
)

func TestParseRef(t *testing.T) {
	for _, tc := range []struct {
		in, block, value string
		ok               bool
	}{
		{"objectives", "objectives", "", true},
		{"scope.included", "scope.included", "", true},
		{"responsibilities.client", "responsibilities", "client", true},
		{"governance.contacts.client", "governance.contacts", "client", true},
		{"milestones.closeout", "milestones", "closeout", true},
		{"signatures.consultant", "signatures", "consultant", true},
		{"responsibilities.boss", "", "", false},
		{"milestones.party", "", "", false},
		{"objectives.client", "", "", false},
		{"nothing", "", "", false},
	} {
		r, err := ParseRef(tc.in)
		if (err == nil) != tc.ok {
			t.Errorf("%s: err = %v, want ok %v", tc.in, err, tc.ok)
			continue
		}
		if tc.ok && (r.Block.Name != tc.block || r.Value != tc.value || r.String() != tc.in) {
			t.Errorf("%s: got %s / %s", tc.in, r.Block.Name, r.Value)
		}
	}
}

func ref(t *testing.T, s string) Ref {
	t.Helper()
	r, err := ParseRef(s)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestOverlaps(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"scope", "scope.included", true},
		{"scope.included", "scope.excluded", false},
		{"responsibilities", "responsibilities.client", true},
		{"responsibilities.client", "responsibilities.consultant", false},
		{"deliverables.list", "deliverables", true},
		{"phases", "milestones", false},
	} {
		if got := Overlaps(ref(t, tc.a), ref(t, tc.b)); got != tc.want {
			t.Errorf("Overlaps(%s, %s) = %v", tc.a, tc.b, got)
		}
	}
}

func TestBuiltinsResolve(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range BuiltinNames() {
		o, ok := Builtin(name)
		if !ok {
			t.Fatalf("builtin %s missing", name)
		}
		for _, s := range o.Sections {
			for _, f := range s.From {
				r := ref(t, f)
				if name == NameFull {
					seen[r.Block.Name] = true
				}
			}
		}
	}
	// The full layout prints every top-level block.
	for _, b := range Blocks() {
		if b.Parent == "" && !seen[b.Name] {
			t.Errorf("full layout does not print %s", b.Name)
		}
	}
}

func TestFilters(t *testing.T) {
	a := &engagement.Agreement{
		Parties:          []engagement.Party{{ID: "c", Role: engagement.RoleConsultant}, {ID: "k", Role: engagement.RoleClient}},
		Responsibilities: []engagement.Responsibility{{ID: "R1", Party: "c"}, {ID: "R2", Party: "k"}},
		Milestones:       []engagement.Milestone{{ID: "M1", Kind: engagement.MilestoneKickoff}, {ID: "M2"}},
	}
	if got := Responsibilities(a, "client"); len(got) != 1 || got[0].ID != "R2" {
		t.Errorf("client responsibilities = %v", got)
	}
	if got := Milestones(a, "kickoff"); len(got) != 1 || got[0].ID != "M1" {
		t.Errorf("kickoff milestones = %v", got)
	}
	if HasContent(a, ref(t, "milestones.closeout")) || !HasContent(a, ref(t, "milestones")) {
		t.Error("milestone content wrong")
	}
	if HasContent(a, ref(t, "governance.contacts")) {
		t.Error("no governance should mean no contacts")
	}
}
