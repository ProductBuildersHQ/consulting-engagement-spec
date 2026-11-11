package render

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/internal/testfixture"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagementio"
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
