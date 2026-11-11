package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/internal/testfixture"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestCommands(t *testing.T) {
	ex := testfixture.Nonprofit()
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"validate", ex}, "valid"},
		{[]string{"lint", ex}, "0 errors, 0 warnings"},
		{[]string{"fmt", "--check", ex}, ""},
		{[]string{"digest", ex}, "sha256:"},
		{[]string{"render", "--flavor", "gfm", ex}, "## 19. Signatures"},
		{[]string{"schema"}, "agreement"},
		{[]string{"schema", "agreement"}, `"$id"`},
	}
	for _, tt := range tests {
		out, err := run(t, tt.args...)
		if err != nil {
			t.Fatalf("%v: %v\n%s", tt.args, err, out)
		}
		if !strings.Contains(out, tt.want) {
			t.Errorf("%v: output does not contain %q:\n%s", tt.args, tt.want, out)
		}
	}
}

func TestExportNeedsOut(t *testing.T) {
	if _, err := run(t, "export", testfixture.Nonprofit()); err == nil {
		t.Fatal("want an error without --out")
	}
	if _, err := run(t, "export", testfixture.Nonprofit(), "--out", "x.html"); err == nil {
		t.Fatal("want an error for an unknown target")
	}
}
