package engagementio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/schema"
)

// ParseAgreement parses an agreement in the given format. The input is
// converted to JSON, validated against the embedded schema, and decoded
// strictly (unknown fields are errors). A failure wraps ErrInvalidDocument;
// schema failures also wrap the validator's error, whose output lists
// JSON-pointer locations.
func ParseAgreement(data []byte, f Format) (*engagement.Agreement, error) {
	var a engagement.Agreement
	if err := parse(data, f, schema.Agreement, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// ParseOutline parses an outline in the given format, as ParseAgreement
// does an agreement.
func ParseOutline(data []byte, f Format) (*engagement.Outline, error) {
	var o engagement.Outline
	if err := parse(data, f, schema.Outline, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func parse(data []byte, f Format, schemaName string, v any) error {
	j, err := ToJSON(data, f)
	if err != nil {
		return err
	}
	if err := schema.Validate(schemaName, j); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDocument, err)
	}
	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDocument, err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: trailing data after document", ErrInvalidDocument)
	}
	return nil
}

// LoadAgreement reads and parses an agreement file (.json, .yaml, or .yml).
func LoadAgreement(path string) (*engagement.Agreement, error) {
	data, f, err := readFile(path)
	if err != nil {
		return nil, err
	}
	a, err := ParseAgreement(data, f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return a, nil
}

// LoadOutline reads and parses an outline file (.json, .yaml, or .yml).
func LoadOutline(path string) (*engagement.Outline, error) {
	data, f, err := readFile(path)
	if err != nil {
		return nil, err
	}
	o, err := ParseOutline(data, f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return o, nil
}

func readFile(path string) ([]byte, Format, error) {
	f, err := FormatFromPath(path)
	if err != nil {
		return nil, "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}
	if info.Size() > MaxFileBytes {
		return nil, "", fmt.Errorf("%w: %s is %d bytes (limit %d)", ErrTooLarge, path, info.Size(), MaxFileBytes)
	}
	data, err := os.ReadFile(path) //nolint:gosec // reading a user-named document is the purpose of this function
	if err != nil {
		return nil, "", err
	}
	return data, f, nil
}

// LoadDocument reads an agreement or an outline, chosen by the document's
// kind, and returns *engagement.Agreement or *engagement.Outline.
func LoadDocument(path string) (any, error) {
	data, f, err := readFile(path)
	if err != nil {
		return nil, err
	}
	j, err := ToJSON(data, f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var head struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(j, &head); err != nil {
		return nil, fmt.Errorf("%s: %w: %w", path, ErrInvalidDocument, err)
	}
	var v any
	if engagement.OutlineKind(head.Kind).Valid() {
		v, err = ParseOutline(data, f)
	} else {
		v, err = ParseAgreement(data, f)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return v, nil
}
