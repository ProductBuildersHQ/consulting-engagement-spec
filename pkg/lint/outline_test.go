package lint

import (
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/outline"
)

func TestOutlineRules(t *testing.T) {
	for _, name := range outline.BuiltinNames() {
		o, _ := outline.Builtin(name)
		if res := Outline(o); len(res.Findings) != 0 {
			t.Errorf("%s: %v", name, res.Findings)
		}
	}
	o := &engagement.Outline{Sections: []engagement.OutlineSection{
		{Heading: "A", From: []string{"objectives", "nope"}},
		{Heading: "A", From: []string{"risks"}},
	}}
	got := rules(Outline(o))
	for _, want := range []string{RuleUnknownBlock, RuleDuplicateHeading} {
		if !slices.Contains(got, want) {
			t.Errorf("%s not reported; got %v", want, got)
		}
	}
}

func TestLayoutEmptySection(t *testing.T) {
	a := load(t)
	o := &engagement.Outline{Sections: []engagement.OutlineSection{
		{Heading: "Hypotheses", From: []string{"hypotheses"}},
		{Heading: "Objectives", From: []string{"objectives"}},
	}}
	res := Layout(a, o)
	if len(res.Findings) != 1 || res.Findings[0].Rule != RuleEmptySection || res.Findings[0].Path != "sections/0" {
		t.Fatalf("findings: %v", res.Findings)
	}
}
