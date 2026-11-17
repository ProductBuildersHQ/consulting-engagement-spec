package crosswalk

import (
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/internal/testfixture"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagementio"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/outline"
)

func TestBuild(t *testing.T) {
	a, err := engagementio.LoadAgreement(testfixture.Nonprofit())
	if err != nil {
		t.Fatal(err)
	}
	target := &engagement.Outline{Title: "Short form", Sections: []engagement.OutlineSection{
		{Heading: "Goal", From: []string{"summary", "objectives"}},
		{Heading: "What the client does", From: []string{"responsibilities.client"}},
		{Heading: "Tests", From: []string{"hypotheses"}},
		{Heading: "Timeline", From: []string{"phases", "milestones"}},
	}}
	cw, err := Build(a, outline.Full(), target)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Row{}
	for _, r := range cw.Rows {
		got[r.Heading] = r
	}
	if r := got["Goal"]; !r.Answered || len(r.Sections) != 2 || r.Sections[0].Label() != "§1 Summary" {
		t.Errorf("Goal: %+v", r)
	}
	if r := got["What the client does"]; !r.Answered || len(r.Sections) != 1 || r.Sections[0].Heading != "Responsibilities" {
		t.Errorf("client: %+v", r)
	}
	if r := got["Tests"]; r.Answered {
		t.Errorf("the example has no hypotheses: %+v", r)
	}
	if r := got["Timeline"]; len(r.Sections) != 2 {
		t.Errorf("Timeline: %+v", r)
	}
	beyond := map[string]bool{}
	for _, s := range cw.Beyond {
		beyond[s.Heading] = true
	}
	if !beyond["Risks"] || beyond["Objectives"] {
		t.Errorf("beyond: %v", cw.Beyond)
	}
	md := cw.Markdown()
	if !strings.Contains(md, "| Tests | Not addressed |") || !strings.Contains(md, "beyond the Short form") {
		t.Errorf("markdown:\n%s", md)
	}
}
