package engagementio

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
)

// DigestPrefix labels a content digest with its algorithm.
const DigestPrefix = "sha256:"

// Canonical serializes v in canonical form: struct field order, two-space
// indentation, no HTML escaping, and a trailing newline. Formatting
// canonical output again yields identical bytes.
func Canonical(v any, f Format) ([]byte, error) {
	j, err := canonicalJSON(v)
	if err != nil {
		return nil, err
	}
	switch f {
	case FormatJSON:
		return j, nil
	case FormatYAML:
		return jsonToYAML(j)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedFormat, string(f))
	}
}

// ContentDigest returns "sha256:<hex>" over the agreed content of a: its
// canonical JSON with the status and the signing records (signed date and
// method) cleared. The digest is the same in every input format, and it
// does not change when a version moves from draft to agreed or when its
// signers sign, so the digest printed on the document a signer reads still
// matches the document once signed. Any change to the terms changes it.
func ContentDigest(a *engagement.Agreement) (string, error) {
	c := *a
	c.Status = ""
	if len(a.Signatures) > 0 {
		c.Signatures = make([]engagement.Signature, len(a.Signatures))
		for i, s := range a.Signatures {
			c.Signatures[i] = engagement.Signature{Party: s.Party, Signer: s.Signer}
		}
	}
	j, err := canonicalJSON(&c)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(j)
	return DigestPrefix + hex.EncodeToString(sum[:]), nil
}

func canonicalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("encode canonical JSON: %w", err)
	}
	return buf.Bytes(), nil
}

// jsonToYAML renders JSON as block-style YAML, preserving key order. The
// JSON text is parsed as a YAML node tree (JSON is a YAML subset), then
// styles are cleared so the encoder picks plain scalars, quoting only where
// needed (for example the string "1.0" or "2026-01-01"). Multi-line strings
// use literal block style so Markdown stays readable.
func jsonToYAML(j []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(j, &doc); err != nil {
		return nil, fmt.Errorf("parse JSON as YAML: %w", err)
	}
	clearStyles(&doc)
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return nil, fmt.Errorf("encode YAML: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("encode YAML: %w", err)
	}
	return buf.Bytes(), nil
}

func clearStyles(n *yaml.Node) {
	n.Style = 0
	n.HeadComment, n.LineComment, n.FootComment = "", "", ""
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" && strings.Contains(n.Value, "\n") {
		n.Style = yaml.LiteralStyle
	}
	for _, c := range n.Content {
		clearStyles(c)
	}
}
