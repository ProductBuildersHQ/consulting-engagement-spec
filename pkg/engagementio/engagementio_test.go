package engagementio

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/internal/testfixture"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
)

func load(t *testing.T) *engagement.Agreement {
	t.Helper()
	a, err := LoadAgreement(testfixture.Nonprofit())
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestExampleIsCanonical(t *testing.T) {
	a := load(t)
	want, err := Canonical(a, FormatYAML)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(testfixture.Nonprofit())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("example is not in canonical form; run cespec fmt")
	}
}

func TestDigestSameForYAMLAndJSON(t *testing.T) {
	a := load(t)
	j, err := Canonical(a, FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ParseAgreement(j, FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	da, err := ContentDigest(a)
	if err != nil {
		t.Fatal(err)
	}
	db, err := ContentDigest(b)
	if err != nil {
		t.Fatal(err)
	}
	if da != db || !strings.HasPrefix(da, DigestPrefix) || len(da) != len(DigestPrefix)+64 {
		t.Fatalf("digests differ or malformed: %s vs %s", da, db)
	}
}

func TestDigestIgnoresStatusAndSigning(t *testing.T) {
	a := load(t)
	before, err := ContentDigest(a)
	if err != nil {
		t.Fatal(err)
	}
	a.Status = engagement.StatusAgreed
	a.Signatures[0].SignedDate = "2026-03-10"
	a.Signatures[0].Method = engagement.MethodElectronic
	after, err := ContentDigest(a)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("signing changed the content digest")
	}
	a.Scope.Excluded = a.Scope.Excluded[1:]
	changed, err := ContentDigest(a)
	if err != nil {
		t.Fatal(err)
	}
	if changed == before {
		t.Fatal("changing the scope did not change the digest")
	}
}

func TestDigestDoesNotMutate(t *testing.T) {
	a := load(t)
	a.Signatures[0].SignedDate = "2026-03-10"
	if _, err := ContentDigest(a); err != nil {
		t.Fatal(err)
	}
	if a.Signatures[0].SignedDate != "2026-03-10" || a.Status == "" {
		t.Fatal("ContentDigest modified its argument")
	}
}

func TestParseRejects(t *testing.T) {
	data, err := os.ReadFile(testfixture.Nonprofit())
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"unknown field":    string(data) + "surprise: true\n",
		"unquoted version": strings.Replace(string(data), `version: "1.0"`, "version: 1.0", 1),
		"bad kind":         strings.Replace(string(data), "kind: proposal", "kind: memo", 1),
		"two documents":    string(data) + "---\nkind: proposal\n",
	}
	for name, in := range cases {
		if _, err := ParseAgreement([]byte(in), FormatYAML); !errors.Is(err, ErrInvalidDocument) {
			t.Errorf("%s: error = %v, want ErrInvalidDocument", name, err)
		}
	}
}
