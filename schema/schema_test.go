package schema

import (
	"bytes"
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/schemagen"
)

func TestNamesAndGet(t *testing.T) {
	if got, want := Names(), []string{Agreement}; !slices.Equal(got, want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
	if _, err := Compiled(Agreement); err != nil {
		t.Fatal(err)
	}
	if _, err := Get("nope"); !errors.Is(err, ErrUnknownSchema) {
		t.Fatalf("Get(nope) error = %v", err)
	}
}

// TestNoDrift fails when the committed schema differs from what the Go
// types generate. Fix it with: go generate ./schema/...
func TestNoDrift(t *testing.T) {
	files, err := schemagen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		got, err := os.ReadFile(f.Name)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, f.Data) {
			t.Errorf("%s is out of date; run go generate ./schema/...", f.Name)
		}
	}
}

func TestValidateRejects(t *testing.T) {
	if err := Validate(Agreement, []byte(`{"schemaVersion":"0.1","kind":"memo"}`)); err == nil {
		t.Fatal("want a validation error")
	}
}
