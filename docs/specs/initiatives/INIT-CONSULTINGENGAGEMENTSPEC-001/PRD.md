# PRD — Consulting Engagement Specification — structured agreements rendered to DOCX and signable PDF

**Initiative:** `INIT-CONSULTINGENGAGEMENTSPEC-001`
**Repository:** `github.com/ProductBuildersHQ/consulting-engagement-spec`
**Workflow:** pbhq-lite
**Status:** Draft — pending stakeholder review (pbhq-lite gate: after PRD)

## 1. Problem

A consulting engagement starts with a proposal or a statement of work (SOW):
the background, what the work covers, what it does not, what each party will
do, what gets delivered, and who signs. Today that document is almost always a
Word file built from a template. That causes three failures:

1. **Nothing is checkable.** Scope, exclusions, deliverables, and
   responsibilities sit in prose. No tool can tell that nothing is listed as
   out of scope, that the client has no responsibilities, that a deliverable
   has no acceptance criteria, or that a milestone points at a deliverable
   that no longer exists. These gaps are the usual source of consulting
   disputes.
2. **No reliable path to a signed document.** Each version is exported by
   hand. The PDF a client signs cannot be traced back to an exact source, and
   two exports of the same text are not the same file.
3. **Not usable by tools or agents.** An AI agent can draft prose, but it
   cannot validate, compare, or revise an agreement it cannot parse. Once
   signed, the agreed deliverables and responsibilities cannot feed the
   project plan without retyping them.

## 2. Goals

- **G-1.** One structured document format for a consulting agreement, with
  stable IDs for every deliverable, milestone, responsibility, exclusion, and
  term.
- **G-2.** Checks a schema cannot express: references resolve, dates are in
  order, and the common gaps are flagged.
- **G-3.** Deterministic output: the same agreement renders to the same
  Markdown, DOCX, and PDF bytes.
- **G-4.** A document the parties can sign, tied to its source by a content
  digest printed on it.
- **G-5.** A model that later lifecycle documents (change orders, acceptance,
  closure) and planning tools can reference by ID.

## 3. Users

| User | Need |
|---|---|
| Independent consultants | A quick, professional proposal that states scope and exclusions clearly |
| Consulting firms | Consistent SOWs across teams, reviewable in version control |
| Volunteer and pro bono teams, such as volunteer consulting programs | A clear working agreement with a nonprofit client, without legal staff or paid tools |
| AI agents | A format they can draft, validate, lint, and revise without a document editor |
| Clients | A readable document that says what they get, what they must provide, and how to sign |

## 4. Lifecycle

A consulting engagement moves through seven stages:

1. **Proposal** — background, objectives, proposed scope, fees.
2. **Agreement** — SOW: scope and exclusions, deliverables, responsibilities,
   acceptance, signatures.
3. **Kickoff** — stakeholders, plan, cadence.
4. **Execution** — progress against deliverables and milestones.
5. **Change** — change requests and amended scope.
6. **Acceptance** — formal acceptance of deliverables.
7. **Closure** — handoff and completion.

**Version 0 covers stages 1 and 2 with one document kind, the Agreement**
(`kind: proposal` or `kind: statement_of_work`). Later stages are separate
documents that reference an agreement by ID and version (see ROADMAP Phase
3). A signed agreement is never edited; a change produces a new version or a
change order.

## 5. Requirements

### Model

- **PRD-FR-101** An agreement records its kind, ID, version, status, title,
  issue date, optional validity date, engagement period, and the version it
  supersedes.
- **PRD-FR-102** An agreement names at least two parties, each with a role
  (consultant, client, sponsor, partner), and the people involved, each
  belonging to a party.
- **PRD-FR-103** An agreement states background, objectives, and scope, with
  included and excluded items listed separately.
- **PRD-FR-104** Deliverables, milestones, responsibilities, assumptions,
  risks, success criteria, terms, and open questions are lists of elements
  with stable IDs.
- **PRD-FR-105** Deliverables reference their phase and the objectives they
  serve; milestones reference the deliverables they review.
- **PRD-FR-106** Commercials state the fee model (pro bono, fixed fee, time
  and materials, retainer), currency, priced lines, expenses, and payment
  terms. Amounts are exact decimals, never floating point.
- **PRD-FR-107** Signature blocks name the signer and party; once signed,
  they record the date and method (wet ink, electronic, digital).
- **PRD-FR-108** Narrative fields hold Markdown.

### Schema and loading

- **PRD-FR-201** A JSON Schema is generated from the Go types and embedded in
  the module.
- **PRD-FR-202** Agreements can be written in YAML or JSON; both validate and
  load identically, and unknown fields are rejected.
- **PRD-FR-203** A canonical form exists, and formatting is idempotent.

### Checks

- **PRD-FR-301** Lint reports errors for duplicate IDs, missing consultant or
  client, unresolved references, dates out of order, missing currency, and an
  agreed document with open questions.
- **PRD-FR-302** Lint warns when nothing is out of scope, the client has no
  responsibilities, a deliverable has no acceptance criteria, an objective is
  served by no deliverable, or there is nowhere to sign.

### Output

- **PRD-FR-401** Render an agreement as Markdown: Pandoc-flavored for
  conversion, GitHub-flavored for reading in a repository.
- **PRD-FR-402** Export to DOCX and PDF through Pandoc, with embedded
  timestamps pinned to the agreement's date.
- **PRD-FR-403** The rendered document prints a content digest that stays the
  same when the agreement is signed and changes when its terms change.
- **PRD-FR-404** The rendered document ends with signature blocks usable for
  wet ink, PDF fill-and-sign tools, or an e-signature service.
- **PRD-FR-405** (Phase 2) Export a PDF with real signature fields, record
  the hash of the frozen PDF, and support cryptographic signing through an
  external service.

### Interfaces

- **PRD-FR-501** All logic lives in importable Go packages; the `cespec` CLI
  is a thin layer over them.

## 6. Non-goals

- **Not a legal template.** The format carries whatever terms the parties
  write. It does not supply legal language, and an attorney should review
  terms and the signing process before they are relied on.
- **Not an e-signature platform.** No signer identity verification, consent
  capture, audit trail, or certificate management. Those come from an
  external signing service.
- **Not a project tracker.** The agreement states what was agreed. What is
  actually happening belongs in planning tools that reference it.
- **No hosted service.** The module is a library and a CLI.

## 7. Success metrics

- The first real engagement (a pro bono nonprofit engagement) is fully
  represented, lints with zero errors, and renders to a DOCX and PDF the
  client accepts without hand edits.
- Rendering the same agreement twice with the same toolchain produces
  byte-identical Markdown, DOCX, and PDF.
- The digest printed on a signed PDF matches `cespec digest` on the source.
- Every lint rule has a test, and every example in the repository lints with
  zero errors.
