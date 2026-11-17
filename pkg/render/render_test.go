package render

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/internal/testfixture"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagementio"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/outline"
)

var update = flag.Bool("update", false, "rewrite golden files")

func load(t *testing.T) *engagement.Agreement {
	t.Helper()
	a, err := engagementio.LoadAgreement(testfixture.Nonprofit())
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// TestGolden compares the rendered example with the committed files in
// the example directory. Refresh them with: go test ./pkg/render -update
func TestGolden(t *testing.T) {
	a := load(t)
	for _, tc := range []struct {
		flavor Flavor
		file   string
	}{
		{FlavorGFM, "agreement.md"},
		{FlavorPandoc, "agreement.pandoc.md"},
	} {
		got, err := Markdown(a, Options{Flavor: tc.flavor})
		if err != nil {
			t.Fatal(err)
		}
		path := testfixture.ExamplePath("nonprofit-impact-measurement", tc.file)
		if *update {
			if err := os.WriteFile(path, got, 0o644); err != nil { //nolint:gosec // golden files are ordinary files
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path) //nolint:gosec // test reads its own golden file
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("%s is out of date; run go test ./pkg/render -update", tc.file)
		}
	}
}

func TestDeterministic(t *testing.T) {
	a := load(t)
	first, err := Markdown(a, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for range 5 {
		again, err := Markdown(a, Options{})
		if err != nil {
			t.Fatal(err)
		}
		if string(again) != string(first) {
			t.Fatal("output differs between runs")
		}
	}
}

func TestSignedAndGFM(t *testing.T) {
	a := load(t)
	a.Signatures[1].SignedDate = "2026-03-10"
	a.Signatures[1].Method = engagement.MethodWetInk
	got, err := Markdown(a, Options{Flavor: FlavorGFM, OmitDigest: true})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"Signed on March 10, 2026 (wet ink)", "# Impact Measurement for Riverbend Food Network"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, unwanted := range []string{"```{=latex}", "sha256:", "header-includes"} {
		if strings.Contains(s, unwanted) {
			t.Errorf("unexpected %q", unwanted)
		}
	}
}

func TestFees(t *testing.T) {
	a := load(t)
	a.Commercials = engagement.Commercials{
		Model: engagement.ModelFixedFee, Currency: "USD",
		Fees: []engagement.Fee{{Description: "Phase 1", Amount: "12500"}, {Description: "Phase 2", Amount: "1234567.50"}},
	}
	got, err := Markdown(a, Options{Flavor: FlavorGFM})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"| Phase 1 |  | 12,500.00 |", "| Phase 2 |  | 1,234,567.50 |", "| **Total** | | **1,247,067.50** |"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestUnknownFlavor(t *testing.T) {
	if _, err := Markdown(load(t), Options{Flavor: "html"}); err == nil {
		t.Fatal("want an error")
	}
}

func render(t *testing.T, a *engagement.Agreement, opts Options) string {
	t.Helper()
	got, err := Markdown(a, opts)
	if err != nil {
		t.Fatal(err)
	}
	return string(got)
}

func TestFullOutlineIsDefault(t *testing.T) {
	a := load(t)
	if render(t, a, Options{}) != render(t, a, Options{Outline: outline.Full()}) {
		t.Fatal("the full outline differs from the default layout")
	}
}

func TestAlignmentFields(t *testing.T) {
	a := load(t)
	a.Hypotheses = []engagement.Hypothesis{{ID: "H1", Statement: "Pantries already record enough", Test: "Compare the sign-in sheets with the funder reports.", Deliverables: []string{"D1"}, Milestone: "M1"}}
	a.Milestones[0].Kind = engagement.MilestoneDecision
	a.Governance.Contacts = []engagement.Contact{{Party: "riverbend", Person: "jordan-park", Role: "Day-to-day lead"}}
	a.Closeout = &engagement.Closeout{
		Evaluations: []engagement.Evaluation{{ID: "E1", Party: "riverbend", Subject: "the engagement and the team", Method: "written evaluation", Due: "2026-07-31"}},
		Reports:     []engagement.Report{{ID: "CR1", To: "northfield", Description: "Written recommendations", Due: "2026-07-31"}},
	}
	a.Signatures[1].Capacity = "Board representative"
	s := render(t, a, Options{Flavor: FlavorGFM})
	for _, want := range []string{
		"## 4. Hypotheses",
		"- **H1. Pantries already record enough.** Compare the sign-in sheets with the funder reports. *Evidence:* D1. *Judged at:* Measure decision, April 24, 2026.",
		"| M1 | Measure decision | Decision | April 24, 2026 |",
		"### Points of contact",
		"| Riverbend Food Network | Jordan Park | Day-to-day lead |",
		". Closeout\n\n**Evaluations**",
		"- **E1.** Riverbend Food Network evaluates the engagement and the team (written evaluation), by July 31, 2026.",
		"- **CR1.** Written recommendations. To Northfield Volunteer Consultants, by July 31, 2026.",
		"Title: Board representative",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q", want)
		}
	}
}

// TestOutlineGolden compares the example in other layouts with the
// committed files. Refresh them with: go test ./pkg/render -update
func TestOutlineGolden(t *testing.T) {
	a := load(t)
	wp, err := engagementio.LoadOutline(testfixture.WorkPlan())
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		file string
		opts Options
	}{
		{"agreement.summary.md", Options{Flavor: FlavorGFM, Outline: outline.Summary()}},
		{"agreement.work-plan.md", Options{Flavor: FlavorGFM, Outline: wp}},
		{"agreement.crosswalk.md", Options{Flavor: FlavorGFM, Crosswalk: wp}},
	} {
		got := render(t, a, tc.opts)
		path := testfixture.ExamplePath("nonprofit-impact-measurement", tc.file)
		if *update {
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil { //nolint:gosec // golden files are ordinary files
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path) //nolint:gosec // test reads its own golden file
		if err != nil {
			t.Fatal(err)
		}
		if got != string(want) {
			t.Errorf("%s is out of date; run go test ./pkg/render -update", tc.file)
		}
	}
}

func TestOutlineOptions(t *testing.T) {
	a := load(t)
	a.Hypotheses = nil
	no := false
	o := &engagement.Outline{
		ID: "plan", Title: "Short plan", Numbered: &no, Information: &no,
		DatePrecision: engagement.PrecisionMonth,
		Sections: []engagement.OutlineSection{
			{Heading: "Timeline", From: []string{"phases", "milestones.decision"}, Note: "Dates are estimates."},
			{Heading: "What we test", From: []string{"hypotheses"}},
			{Heading: "Client commitments", From: []string{"responsibilities.client"}},
		},
	}
	s := render(t, a, Options{Flavor: FlavorGFM, Outline: o})
	for _, want := range []string{"## Timeline\n\nDates are estimates.", "| Discover | March – April 2026 |", "## Client commitments\n\n- Provide"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, unwanted := range []string{"| **Document** |", "What we test", "### Milestones", "## 1."} {
		if strings.Contains(s, unwanted) {
			t.Errorf("unexpected %q", unwanted)
		}
	}
	a.Milestones[0].Kind = engagement.MilestoneDecision
	if s := render(t, a, Options{Flavor: FlavorGFM, Outline: o}); !strings.Contains(s, "### Milestones") {
		t.Error("a decision milestone should print")
	}
}

func TestDigestSameInEveryOutline(t *testing.T) {
	a := load(t)
	d, err := engagementio.ContentDigest(a)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range []*engagement.Outline{outline.Full(), outline.Summary()} {
		if !strings.Contains(render(t, a, Options{Outline: o}), d) {
			t.Errorf("%s: digest %s missing", o.ID, d)
		}
	}
}

func TestUnknownBlock(t *testing.T) {
	o := &engagement.Outline{ID: "x", Sections: []engagement.OutlineSection{{Heading: "A", From: []string{"nope"}}}}
	if _, err := Markdown(load(t), Options{Outline: o}); err == nil {
		t.Fatal("want an error")
	}
}
