// Package engagement defines the consulting engagement data model. Go
// structs are the source of truth; the JSON Schemas under schema/ are
// generated from them.
//
// The first document kind is the Agreement: a proposal or statement of work
// that states what a consultant and a client agree to, including what is out
// of scope, what each party is responsible for, and who signs.
package engagement

// SchemaVersion is the version of the document format this package reads
// and writes.
const SchemaVersion = "0.1"

// Agreement is a proposal or statement of work between a consultant and a
// client. Narrative fields hold Markdown; everything a tool needs to check,
// compare, or track is a structured field with a stable ID.
type Agreement struct {
	SchemaVersion string `json:"schemaVersion" jsonschema:"required,enum=0.1"`
	Kind          Kind   `json:"kind" jsonschema:"required"`
	// ID is a stable identifier for the agreement across versions.
	ID string `json:"id" jsonschema:"required,pattern=^[a-z0-9][a-z0-9-]*$"`
	// Version is MAJOR.MINOR. A new version is a new document; a signed
	// version is never edited.
	Version string `json:"version" jsonschema:"required,pattern=^[0-9]+\\.[0-9]+$"`
	Status  Status `json:"status" jsonschema:"required"`
	Title   string `json:"title" jsonschema:"required,minLength=1"`
	// Subtitle is an optional second line under the title.
	Subtitle string `json:"subtitle,omitempty"`
	// Date is the issue date of this version (YYYY-MM-DD).
	Date string `json:"date" jsonschema:"required,format=date"`
	// ValidUntil is the date after which a proposal lapses unless accepted.
	ValidUntil string `json:"validUntil,omitempty" jsonschema:"format=date"`
	// Period is when the engagement runs.
	Period *Period `json:"period,omitempty"`
	// Supersedes names the version this one replaces.
	Supersedes *VersionRef `json:"supersedes,omitempty"`

	Parties []Party  `json:"parties" jsonschema:"required,minItems=2"`
	People  []Person `json:"people,omitempty"`

	// Summary is a short executive summary (Markdown).
	Summary string `json:"summary,omitempty"`
	// Background describes the client's situation and why the work is
	// needed now (Markdown).
	Background string      `json:"background" jsonschema:"required,minLength=1"`
	Objectives []Objective `json:"objectives" jsonschema:"required,minItems=1"`
	Scope      Scope       `json:"scope" jsonschema:"required"`
	// Approach describes the method and how the work is organized
	// (Markdown).
	Approach         string           `json:"approach,omitempty"`
	Phases           []Phase          `json:"phases,omitempty"`
	Deliverables     []Deliverable    `json:"deliverables" jsonschema:"required,minItems=1"`
	Milestones       []Milestone      `json:"milestones,omitempty"`
	Responsibilities []Responsibility `json:"responsibilities" jsonschema:"required,minItems=1"`
	Governance       *Governance      `json:"governance,omitempty"`
	Assumptions      []Assumption     `json:"assumptions,omitempty"`
	Risks            []Risk           `json:"risks,omitempty"`
	SuccessCriteria  []Criterion      `json:"successCriteria,omitempty"`
	Acceptance       *Acceptance      `json:"acceptance,omitempty"`
	Commercials      Commercials      `json:"commercials" jsonschema:"required"`
	// ChangeControl describes how scope, schedule, or fees change after
	// agreement (Markdown).
	ChangeControl string `json:"changeControl,omitempty"`
	Terms         []Term `json:"terms,omitempty"`
	// OpenQuestions are items still to settle. An agreed document has none.
	OpenQuestions []OpenQuestion `json:"openQuestions,omitempty"`
	Signatures    []Signature    `json:"signatures,omitempty"`
}

// Period is a date range (YYYY-MM-DD, inclusive).
type Period struct {
	Start string `json:"start" jsonschema:"required,format=date"`
	End   string `json:"end" jsonschema:"required,format=date"`
}

// VersionRef names one version of an agreement.
type VersionRef struct {
	ID      string `json:"id" jsonschema:"required,pattern=^[a-z0-9][a-z0-9-]*$"`
	Version string `json:"version" jsonschema:"required,pattern=^[0-9]+\\.[0-9]+$"`
}

// Party is an organization that is party to the agreement.
type Party struct {
	ID   string    `json:"id" jsonschema:"required,pattern=^[a-z0-9][a-z0-9-]*$"`
	Role PartyRole `json:"role" jsonschema:"required"`
	Name string    `json:"name" jsonschema:"required,minLength=1"`
	// LegalName is the registered name, when it differs from Name.
	LegalName   string `json:"legalName,omitempty"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty" jsonschema:"format=uri"`
	Address     string `json:"address,omitempty"`
}

// Person is someone named in the agreement: a team member, a contact, or a
// signer.
type Person struct {
	ID    string `json:"id" jsonschema:"required,pattern=^[a-z0-9][a-z0-9-]*$"`
	Name  string `json:"name" jsonschema:"required,minLength=1"`
	Party string `json:"party" jsonschema:"required"`
	// Title is the person's job title at their organization.
	Title string `json:"title,omitempty"`
	// Role is the person's role in this engagement, such as "Project
	// Leader" or "Day-to-day contact".
	Role  string `json:"role,omitempty"`
	Email string `json:"email,omitempty" jsonschema:"format=email"`
}

// Objective is an outcome the engagement is meant to achieve.
type Objective struct {
	ID          string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Title       string `json:"title" jsonschema:"required,minLength=1"`
	Description string `json:"description,omitempty"`
}

// Scope states what the engagement covers and, as importantly, what it
// does not.
type Scope struct {
	Summary  string      `json:"summary,omitempty"`
	Included []ScopeItem `json:"included" jsonschema:"required,minItems=1"`
	Excluded []ScopeItem `json:"excluded,omitempty"`
}

// ScopeItem is one included or excluded piece of work.
type ScopeItem struct {
	ID          string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Title       string `json:"title" jsonschema:"required,minLength=1"`
	Description string `json:"description,omitempty"`
}

// Phase is a stage of the work.
type Phase struct {
	ID    string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Title string `json:"title" jsonschema:"required,minLength=1"`
	Start string `json:"start,omitempty" jsonschema:"format=date"`
	End   string `json:"end,omitempty" jsonschema:"format=date"`
	// Goal is what the phase sets out to do (Markdown).
	Goal string `json:"goal,omitempty"`
}

// Deliverable is a tangible output the consultant (or, occasionally, the
// client) commits to produce.
type Deliverable struct {
	ID          string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Title       string `json:"title" jsonschema:"required,minLength=1"`
	Description string `json:"description,omitempty"`
	// Owner is the ID of the party that produces it; the consultant when
	// empty.
	Owner  string `json:"owner,omitempty"`
	Phase  string `json:"phase,omitempty"`
	Due    string `json:"due,omitempty" jsonschema:"format=date"`
	Format string `json:"format,omitempty"`
	// Objectives are the IDs of the objectives the deliverable serves.
	Objectives         []string `json:"objectives,omitempty"`
	AcceptanceCriteria []string `json:"acceptanceCriteria,omitempty"`
}

// Milestone is a dated checkpoint, usually a meeting where something is
// reviewed or decided.
type Milestone struct {
	ID          string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Title       string `json:"title" jsonschema:"required,minLength=1"`
	Date        string `json:"date,omitempty" jsonschema:"format=date"`
	Phase       string `json:"phase,omitempty"`
	Description string `json:"description,omitempty"`
	// Decision is what is decided or approved at the milestone.
	Decision string `json:"decision,omitempty"`
	// Deliverables are the IDs of deliverables presented at the milestone.
	Deliverables []string `json:"deliverables,omitempty"`
}

// Responsibility is something one party commits to do so the work can
// succeed.
type Responsibility struct {
	ID          string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Party       string `json:"party" jsonschema:"required"`
	Description string `json:"description" jsonschema:"required,minLength=1"`
	Due         string `json:"due,omitempty" jsonschema:"format=date"`
}

// Governance describes how the parties work together.
type Governance struct {
	Summary  string    `json:"summary,omitempty"`
	Meetings []Meeting `json:"meetings,omitempty"`
}

// Meeting is a recurring or one-time meeting.
type Meeting struct {
	Title        string `json:"title" jsonschema:"required,minLength=1"`
	Participants string `json:"participants,omitempty"`
	Frequency    string `json:"frequency,omitempty"`
	Duration     string `json:"duration,omitempty"`
}

// Assumption is a condition the plan depends on.
type Assumption struct {
	ID          string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Description string `json:"description" jsonschema:"required,minLength=1"`
}

// Risk is something that could affect delivery, with its mitigation.
type Risk struct {
	ID          string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Description string `json:"description" jsonschema:"required,minLength=1"`
	Mitigation  string `json:"mitigation,omitempty"`
	// Owner is a party ID.
	Owner string `json:"owner,omitempty"`
}

// Criterion is a measure of whether the engagement succeeded.
type Criterion struct {
	ID          string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Description string `json:"description" jsonschema:"required,minLength=1"`
}

// Acceptance describes how deliverables are reviewed and accepted.
type Acceptance struct {
	Process string `json:"process" jsonschema:"required,minLength=1"`
	// ReviewDays is how many business days the client has to review a
	// deliverable before it is treated as accepted. Zero means not set.
	ReviewDays int `json:"reviewDays,omitempty" jsonschema:"minimum=0"`
}

// Commercials states the fee model, fees, and expenses.
type Commercials struct {
	Model CommercialModel `json:"model" jsonschema:"required"`
	// Currency is an ISO 4217 code, required when Fees is not empty.
	Currency string `json:"currency,omitempty" jsonschema:"pattern=^[A-Z]{3}$"`
	Summary  string `json:"summary,omitempty"`
	Fees     []Fee  `json:"fees,omitempty"`
	Expenses string `json:"expenses,omitempty"`
	// PaymentTerms describes invoicing and payment (Markdown).
	PaymentTerms string `json:"paymentTerms,omitempty"`
}

// Fee is one priced line. Amount is a decimal string in major units, such
// as "12500" or "12500.00", so no float rounding reaches the document.
type Fee struct {
	Description string `json:"description" jsonschema:"required,minLength=1"`
	Amount      string `json:"amount" jsonschema:"required,pattern=^[0-9]+(\\.[0-9]{2})?$"`
	Due         string `json:"due,omitempty"`
}

// Term is a contractual or working term, such as confidentiality or
// ownership of work product.
type Term struct {
	ID    string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Title string `json:"title" jsonschema:"required,minLength=1"`
	Text  string `json:"text" jsonschema:"required,minLength=1"`
}

// OpenQuestion is an item the parties still have to settle.
type OpenQuestion struct {
	ID       string `json:"id" jsonschema:"required,pattern=^[A-Za-z0-9][A-Za-z0-9._-]*$"`
	Question string `json:"question" jsonschema:"required,minLength=1"`
	// Owner is a party or person ID.
	Owner string `json:"owner,omitempty"`
	Due   string `json:"due,omitempty" jsonschema:"format=date"`
}

// Signature is a signature block for an authorized signer, and, once
// signed, the record of the signing.
type Signature struct {
	Party string `json:"party" jsonschema:"required"`
	// Signer is a person ID.
	Signer string `json:"signer" jsonschema:"required"`
	// SignedDate is set once the signer has signed (YYYY-MM-DD).
	SignedDate string          `json:"signedDate,omitempty" jsonschema:"format=date"`
	Method     SignatureMethod `json:"method,omitempty"`
}
