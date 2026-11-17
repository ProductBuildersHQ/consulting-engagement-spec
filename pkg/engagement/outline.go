package engagement

// Outline lays an agreement out in a chosen set of sections, headings, and
// order, such as the one-page work plan a program office or a client asks
// for. It holds headings and references to parts of an agreement, never the
// agreement's content, so one agreement renders in any outline.
type Outline struct {
	SchemaVersion string      `json:"schemaVersion" jsonschema:"required,enum=0.1"`
	Kind          OutlineKind `json:"kind" jsonschema:"required"`
	ID            string      `json:"id" jsonschema:"required,pattern=^[a-z0-9][a-z0-9-]*$"`
	// Title names the outline, such as "One-page work plan". The crosswalk
	// and the information table use it.
	Title string `json:"title" jsonschema:"required,minLength=1"`
	// Numbered numbers the sections; true when not set.
	Numbered *bool `json:"numbered,omitempty"`
	// Information prints the information table; true when not set.
	Information *bool `json:"information,omitempty"`
	// DatePrecision is how dates print in this layout; day when not set.
	DatePrecision DatePrecision    `json:"datePrecision,omitempty"`
	Sections      []OutlineSection `json:"sections" jsonschema:"required,minItems=1"`
}

// OutlineSection is one section of an outline.
type OutlineSection struct {
	Heading string `json:"heading" jsonschema:"required,minLength=1"`
	// From names the blocks of the agreement the section prints, in order,
	// such as "hypotheses" or "responsibilities.client".
	From []string `json:"from" jsonschema:"required,minItems=1"`
	// Note is the outline author's own text, printed under the heading
	// (Markdown).
	Note string `json:"note,omitempty"`
}

// IsNumbered reports whether the outline numbers its sections.
func (o *Outline) IsNumbered() bool { return o.Numbered == nil || *o.Numbered }

// ShowsInformation reports whether the outline prints the information table.
func (o *Outline) ShowsInformation() bool { return o.Information == nil || *o.Information }
