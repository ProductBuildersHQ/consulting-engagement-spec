// Package schemagen generates the JSON Schemas for the consulting
// engagement document kinds from the Go types in pkg/engagement.
//
// Go structs are the source of truth. Enumeration types are emitted as
// string enums whose values come from the pkg/engagement All… functions, so
// the schema and the code cannot drift.
package schemagen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/invopop/jsonschema"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
)

// BaseURL is the prefix of every schema $id.
const BaseURL = "https://github.com/ProductBuildersHQ/consulting-engagement-spec/schema/v0/"

// Definition describes one generated schema.
type Definition struct {
	// Name is the file stem: <Name>.schema.json.
	Name        string
	Title       string
	Description string
	Value       any
}

// Definitions returns the schemas this module publishes, in a stable order.
func Definitions() []Definition {
	return []Definition{
		{"agreement", "Consulting Agreement", "A consulting proposal or statement of work: parties, background, objectives, scope and exclusions, deliverables, milestones, responsibilities, commercials, terms, and signatures.", &engagement.Agreement{}},
	}
}

// ID returns the $id URL for a schema name.
func ID(name string) string {
	return BaseURL + name + ".schema.json"
}

// FileName returns the file name for a schema name.
func FileName(name string) string {
	return name + ".schema.json"
}

// File is one generated schema file.
type File struct {
	Name string
	Data []byte
}

// Generate returns every schema as indented JSON with a trailing newline,
// in Definitions order. Output is deterministic.
func Generate() ([]File, error) {
	defs := Definitions()
	files := make([]File, 0, len(defs))
	for _, d := range defs {
		data, err := GenerateOne(d)
		if err != nil {
			return nil, err
		}
		files = append(files, File{Name: FileName(d.Name), Data: data})
	}
	return files, nil
}

// GenerateOne returns one schema as indented JSON with a trailing newline.
func GenerateOne(d Definition) ([]byte, error) {
	r := &jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true,
		Mapper:                     enumMapper,
	}
	s := r.Reflect(d.Value)
	if s == nil {
		return nil, fmt.Errorf("schemagen: reflect %s: nil schema", d.Name)
	}
	s.ID = jsonschema.ID(ID(d.Name))
	s.Title = d.Title
	s.Description = d.Description

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return nil, fmt.Errorf("schemagen: encode %s: %w", d.Name, err)
	}
	return buf.Bytes(), nil
}

// EnumValues maps each pkg/engagement enumeration type to its values.
func EnumValues() map[reflect.Type][]string {
	return map[reflect.Type][]string{
		reflect.TypeFor[engagement.Kind]():            strs(engagement.AllKinds()),
		reflect.TypeFor[engagement.Status]():          strs(engagement.AllStatuses()),
		reflect.TypeFor[engagement.PartyRole]():       strs(engagement.AllPartyRoles()),
		reflect.TypeFor[engagement.CommercialModel](): strs(engagement.AllCommercialModels()),
		reflect.TypeFor[engagement.SignatureMethod](): strs(engagement.AllSignatureMethods()),
	}
}

// enumMapper emits enumeration types as string enums.
func enumMapper(t reflect.Type) *jsonschema.Schema {
	values, ok := EnumValues()[t]
	if !ok {
		return nil
	}
	enum := make([]any, len(values))
	for i, v := range values {
		enum[i] = v
	}
	return &jsonschema.Schema{Type: "string", Enum: enum}
}

func strs[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return slices.Clip(out)
}
