# TRD — Consulting Engagement Specification — structured agreements rendered to DOCX and signable PDF

**Initiative:** `INIT-CONSULTINGENGAGEMENTSPEC-001`
**Repository:** `github.com/ProductBuildersHQ/consulting-engagement-spec`
**Workflow:** pbhq-lite
**Status:** Draft

Requirements below are `TRD-NNN`. The PRD requirements they satisfy are noted
in brackets.

## 1. Packages

| Package | Role |
|---|---|
| `pkg/engagement` | Go types: the source of truth for the format |
| `pkg/schemagen` | Generates JSON Schemas from the types |
| `schema` | Embedded generated schemas: `Names`, `Get`, `Compiled`, `Validate` |
| `pkg/engagementio` | Load (YAML or JSON), canonical form, content digest |
| `pkg/lint` | Semantic checks the schema cannot express |
| `pkg/render` | Deterministic Markdown output |
| `pkg/pandoc` | Builds the Pandoc command for DOCX and PDF; does not run it |
| `internal/cli`, `cmd/cespec` | Thin Cobra CLI over the packages |
| `cmd/generate-schema` | The only writer of `schema/*.schema.json` |

**TRD-001** Library packages do no network I/O and start no processes. Only
the CLI runs Pandoc. [PRD-FR-501]

## 2. Data model

**TRD-002** The format version is `schemaVersion: "0.1"`. JSON property
names are camelCase; enum values are snake_case. [PRD-FR-101]

### Agreement

| Field | Required | Purpose |
|---|---|---|
| `schemaVersion` | yes | Format version, `"0.1"` |
| `kind` | yes | `proposal` or `statement_of_work` |
| `id` | yes | Stable slug across versions (`^[a-z0-9][a-z0-9-]*$`) |
| `version` | yes | `MAJOR.MINOR`, as a string |
| `status` | yes | `draft`, `in_review`, `agreed`, `superseded`, `withdrawn` |
| `title`, `subtitle` | title | Document title and optional second line |
| `date` | yes | Issue date of this version (`YYYY-MM-DD`) |
| `validUntil` | no | Date a proposal lapses unless accepted |
| `period` | no | `start` and `end` of the engagement |
| `supersedes` | no | `id` and `version` of the version this replaces |
| `parties` | yes, ≥2 | Organizations party to the agreement |
| `people` | no | Team members, contacts, and signers |
| `summary` | no | Executive summary (Markdown) |
| `background` | yes | Client situation and why now (Markdown) |
| `objectives` | yes, ≥1 | What the engagement is meant to achieve |
| `hypotheses` | no | What the engagement sets out to test, how, and where it is judged |
| `scope` | yes | `summary`, `included` (≥1), `excluded` |
| `approach` | no | Method (Markdown) |
| `phases` | no | Stages of the work, with optional dates and goal |
| `deliverables` | yes, ≥1 | Outputs, with owner, phase, due date, format, objectives served, acceptance criteria |
| `milestones` | no | Dated checkpoints, with an optional kind, the decision taken, and deliverables reviewed |
| `responsibilities` | yes, ≥1 | What each party commits to do, with optional due date |
| `governance` | no | `summary`, `contacts` (each party's point of contact), and `meetings` (title, participants, frequency, duration) |
| `assumptions` | no | Conditions the plan depends on |
| `risks` | no | Risk, mitigation, owning party |
| `successCriteria` | no | Measures of whether the engagement succeeded |
| `acceptance` | no | Review `process` and `reviewDays` |
| `closeout` | no | `summary`, `evaluations` each party gives, and `reports` owed to a party |
| `commercials` | yes | `model`, `currency`, `summary`, `fees`, `expenses`, `paymentTerms` |
| `changeControl` | no | How scope, schedule, or fees change (Markdown) |
| `terms` | no | Titled terms, such as confidentiality or ownership (Markdown) |
| `openQuestions` | no | Items still to settle, with owner and due date |
| `signatures` | no | Signature blocks and, once signed, the signing record |

### Elements

- **Party:** `id`, `role` (`consultant`, `client`, `sponsor`, `partner`),
  `name`, `legalName`, `description`, `website`, `address`.
- **Person:** `id`, `name`, `party`, `title` (job title), `role` (role in
  this engagement), `email`.
- **Deliverable:** `owner` is a party ID; when empty the first consultant
  party owns it.
- **Fee:** `description`, `amount` (decimal string in major units, such as
  `"12500"` or `"12500.00"`), `due` (free text). [PRD-FR-106]
- **Signature:** `party`, `signer` (person ID), `capacity` (the capacity in
  which the signer signs, such as "Board representative"; printed in place
  of the person's title when set), `signedDate`, `method` (`wet_ink`,
  `electronic`, `digital`). A party may have several signature blocks.
  [PRD-FR-107, PRD-FR-604]
- **Hypothesis:** `id`, `statement` (the claim), `test` (how the work tests
  it, Markdown), `deliverables` (IDs of the deliverables that produce the
  evidence), `milestone` (ID of the milestone where the result is judged).
  [PRD-FR-601]
- **Milestone `kind`:** one of `kickoff`, `checkpoint`, `review`,
  `decision`, `final`, `closeout`. Optional; it labels the milestone and
  lets an outline select by kind. [PRD-FR-602]
- **Contact:** `party`, `person` (person ID), `role` (free text, such as
  "Day-to-day lead"). [PRD-FR-604]
- **Closeout:** `summary` (Markdown); `evaluations`, each with `id`, `party`
  (who evaluates), `subject` (what is evaluated, free text), `method`, and
  `due`; `reports`, each with `id`, `to` (party ID), `description`, and
  `due`. [PRD-FR-603]

**TRD-003** Element IDs of objectives, hypotheses, scope items, phases,
deliverables, milestones, responsibilities, assumptions, risks, criteria,
evaluations, reports, terms, and open questions match `^[A-Za-z0-9][A-Za-z0-9._-]*$`, so authors can use short
labels such as `D1` or `M2`. Party and person IDs are lowercase slugs.
[PRD-FR-104]

**TRD-004** All IDs share one namespace: parties, people, and every element
kind. A reference such as an open question's `owner` (a party or a person)
is therefore never ambiguous. [PRD-FR-104]

**TRD-005** Narrative fields are Markdown. In list items and table cells the
renderer collapses them to one line, so multi-paragraph text belongs in
`summary`, `background`, `approach`, `changeControl`, and `terms`.
[PRD-FR-108]

**TRD-006** The alignment fields (`hypotheses`, milestone `kind`,
`governance.contacts`, `closeout`, signature `capacity`) are optional
additions; `schemaVersion` stays `"0.1"`, and an agreement without them
renders as before, since sections are numbered only when present. [PRD-FR-601 to PRD-FR-604]

## 3. Schema

**TRD-010** Go structs in `pkg/engagement` are the source of truth.
`pkg/schemagen` reflects them with `invopop/jsonschema`, taking required
fields, patterns, formats, and minimums from `jsonschema` struct tags.
[PRD-FR-201]

**TRD-011** Enumeration types are emitted as string enums whose values come
from the `All…` functions (`AllKinds`, `AllStatuses`, `AllPartyRoles`,
`AllCommercialModels`, `AllSignatureMethods`), so code and schema cannot
drift.

**TRD-012** Schemas are regenerated with
`go run ./cmd/generate-schema -out schema` (or `go generate ./schema/...`),
never edited by hand, embedded with `//go:embed`, and pass
`schemakit lint --property-case camelCase`. Each `$id` is
`https://github.com/ProductBuildersHQ/consulting-engagement-spec/schema/v0/<name>.schema.json`.

## 4. Loading and canonical form

**TRD-020** `engagementio.ParseAgreement` converts input to JSON, validates
it against the embedded schema, then decodes it strictly: unknown fields and
trailing data are errors. Failures wrap `ErrInvalidDocument`. [PRD-FR-202]

**TRD-021** YAML input must hold exactly one document with string keys.
Unquoted YAML dates become `YYYY-MM-DD` strings. Numbers stay numbers, so
`version`, `schemaVersion`, and fee `amount` must be quoted in YAML
(`version: "1.0"`); the schema error names the field otherwise.

**TRD-022** Inputs over 10 MiB are rejected with `ErrTooLarge`.

**TRD-023** `engagementio.Canonical` writes struct field order, two-space
indentation, no HTML escaping, and a trailing newline. YAML output is block
style, quotes only where needed, and uses literal blocks for multi-line
strings. Formatting canonical output again yields the same bytes.
[PRD-FR-203]

## 5. Content digest

**TRD-030** `engagementio.ContentDigest` returns `sha256:<hex>` over the
canonical JSON of the agreement with `status` cleared and each signature
reduced to `party` and `signer` (`signedDate` and `method` cleared).
[PRD-FR-403]

Why: the digest identifies the agreed terms. It is the same for YAML and
JSON input. It does not change when a version moves from draft to agreed or
when signers sign, so the digest printed on the document a signer reads
still matches the source afterwards. Any change to the terms, including who
signs, changes it.

## 6. Lint

**TRD-040** `lint.Agreement` returns findings in a fixed order (check order,
then document order). Each finding has a severity, rule ID, path (such as
`deliverables/D3/phase`), and message. Errors make an agreement unfit to
send for signature; the CLI's `render` and `export` refuse to run when there
are any. [PRD-FR-301, PRD-FR-302]

| Rule | Severity | Checks |
|---|---|---|
| `duplicate-id` | error | An ID is used twice in the shared namespace |
| `missing-role` | error | No consultant party, or no client party |
| `unknown-party` | error | A person, responsibility, deliverable owner, risk owner, or signature names an unknown party |
| `unknown-person` | error | A signature's signer is not a person |
| `signer-party` | error | The signer belongs to a different party than the signature block |
| `unknown-phase` | error | A deliverable or milestone names an unknown phase |
| `unknown-objective` | error | A deliverable names an unknown objective |
| `unknown-deliverable` | error | A milestone names an unknown deliverable |
| `unknown-owner` | error | An open question's owner is neither a party nor a person |
| `date-order` | error | An end date is before its start (period, phases, `validUntil` before `date`) |
| `date-outside-period` | warning | A phase date, deliverable due date, or milestone date is outside the engagement period |
| `currency-required` | error | Fees are listed with no currency |
| `pro-bono-fee` | warning | A pro bono engagement lists a non-zero fee |
| `agreed-open-questions` | error | Status is `agreed` but open questions remain |
| `agreed-unsigned` | warning | Status is `agreed` but a signature has no `signedDate` |
| `no-exclusions` | warning | Nothing is out of scope |
| `no-client-responsibility` | warning | No responsibility belongs to a client party |
| `no-acceptance-criteria` | warning | A deliverable has no acceptance criteria and there is no acceptance process |
| `objective-unserved` | warning | Some deliverables link objectives, but this objective has none |
| `no-signatures` | warning | No signature blocks |
| `unknown-reference` | error | A hypothesis names an unknown deliverable or milestone, a contact names an unknown party or person, or a closeout evaluation or report names an unknown party |
| `contact-party` | error | A contact's person belongs to a different party |
| `hypothesis-untested` | warning | A hypothesis has no `test`, no deliverables, and no milestone |
| `no-kickoff`, `no-closeout` | warning | Some milestones have a kind, but none is a kickoff, or none is a closeout |
| `no-client-contact` | warning | Contacts are listed, but none belongs to a client party |

## 7. Rendering

**TRD-050** `render.Markdown` is deterministic: same agreement and options,
same bytes. Map iteration never reaches output. [PRD-FR-401]

**TRD-051** Flavors:

- `pandoc` (default): a YAML metadata block with title, subtitle, date,
  `lang`, page geometry, font size, link colors, and `header-includes` that
  set a PDF footer naming the kind, ID, and version on every page, plus raw
  page breaks for LaTeX and DOCX (`{=latex}` and `{=openxml}` blocks).
- `gfm`: a `#` title and no raw blocks, for reading in a repository.

**TRD-052** The document opens with an information table (kind, version,
status, reference, prepared by and for, sponsor or partner, date, valid
until, period, supersedes, and the first 16 hex digits of the digest). Then
sections, numbered in order, only when present:

1. Summary
2. Background
3. Objectives
4. Hypotheses (statement, how it is tested, where it is judged)
5. Scope (In scope; Out of scope)
6. Approach (with a Phases table)
7. Deliverables (title, phase, due date, non-consultant owner, description,
   format, "Accepted when")
8. Milestones (with the kind, when set)
9. Responsibilities (grouped by party, in party order)
10. Team and contacts
11. Working together (points of contact, governance, and meetings)
12. Assumptions
13. Risks
14. Measures of success
15. Review and acceptance
16. Closeout (evaluations and reports)
17. Fees and expenses (with a total when there is more than one fee)
18. Changes to this agreement
19. Terms
20. Open questions
21. Signatures

This order is the built-in `full` outline (§10).

Dates print as "November 2, 2026". `--no-digest` leaves the digest out.

**TRD-053** Signatures start on a new page. The section states the kind and
version being agreed and the full content digest, then one block per
signature: "For <party>", a signature line, name, title, and a date line.
A signed block prints "Signed on <date> (<method>)" instead of the lines.
[PRD-FR-404]

## 8. Export

**TRD-060** `pandoc.Command` returns arguments and extra environment for one
conversion: `--from=markdown --standalone --output=<out>`, then
`--to=docx` (with an optional `--reference-doc`) or
`--pdf-engine=<engine>` (default `lualatex`), then any extra arguments and
the input. The target comes from the output extension (`.docx`, `.pdf`).
[PRD-FR-402]

**TRD-061** For reproducible files, the environment sets
`SOURCE_DATE_EPOCH` to the agreement's `date` at midnight UTC and
`FORCE_SOURCE_DATE=1`, so Pandoc and TeX embed that date instead of the
current time. The default engine is LuaLaTeX because, unlike XeLaTeX and
pdfLaTeX, it writes the same PDF ID on every run.

**TRD-063** After a DOCX export, `pandoc.NormalizeZip` sets every ZIP entry's
modified time to the agreement's date. Some Pandoc versions stamp files
copied from the built-in reference document with the current time, which
otherwise makes identical exports differ.

**TRD-062** `cespec export` writes the Pandoc Markdown to a temporary
directory (or next to the output with `--keep-markdown`), runs the
configured `pandoc` binary, and removes the temporary directory.

## 9. Signing

**TRD-070** Phase 1 (this version): the PDF has a printed signature block. It
works for wet ink and scanning, for PDF fill-and-sign tools such as Acrobat
Fill & Sign, and for e-signature services where the sender places fields on
the block. Signing is recorded back into the agreement as `signedDate` and
`method`. [PRD-FR-404]

**TRD-071** Phase 2: [PRD-FR-405]

- Export a PDF with AcroForm signature fields, one per signature block, so
  PDF readers offer "click to sign".
- **Freeze:** write a manifest recording the agreement ID, version, content
  digest, toolchain versions, and the SHA-256 of the frozen unsigned PDF.
  Signing changes PDF bytes, so the signed PDF is a separate artifact linked
  to the frozen one.
- Apply cryptographic signatures through an external signing service, using
  PAdES (ETSI EN 319 142), with long-term validation profiles where needed.
- A `verify` command checks a signed PDF against the manifest and the
  source.

**TRD-072** For U.S. agreements, electronic signing should meet ESIGN and
UETA expectations: intent to sign, consent to do business electronically,
attribution, and record retention. These come from the signing service and
the parties' process, not from this module.

## 10. Outlines and crosswalk

An outline lays an agreement out in another set of sections, headings, and
order, such as the one-page work plan a program office or a client asks for,
without a second document. [PRD-FR-605 to PRD-FR-607]

**TRD-080** An outline is its own document kind, with its own generated
schema (`outline.schema.json`), loaded and validated like an agreement:

| Field | Required | Purpose |
|---|---|---|
| `schemaVersion` | yes | `"0.1"` |
| `kind` | yes | `outline` |
| `id` | yes | Slug |
| `title` | yes | The outline's name, used in the crosswalk heading |
| `numbered` | no | Number the sections (default true) |
| `datePrecision` | no | `day` (default) or `month`, for dates in this layout |
| `information` | no | Print the information table (default true) |
| `sections` | yes, ≥1 | Each with `heading`, `from` (≥1 block), optional `note` (Markdown printed under the heading) |

**TRD-081** `from` names blocks: the parts of an agreement the renderer can
print. Each section of the `full` outline is a block (`summary`,
`background`, `objectives`, `hypotheses`, `scope`, `scope.included`,
`scope.excluded`, `approach`, `phases`, `deliverables`, `milestones`,
`responsibilities`, `people`, `governance`, `governance.contacts`,
`assumptions`, `risks`, `successCriteria`, `acceptance`, `closeout`,
`commercials`, `changeControl`, `terms`, `openQuestions`, `signatures`).
Filters narrow a block:

- by party role: `responsibilities.client`, `responsibilities.consultant`,
  `governance.contacts.client`, `signatures.client`;
- by milestone kind: `milestones.kickoff`, `milestones.closeout`, and so on.

Blocks in one section print in the order listed, under the section's
heading, with their own headings one level down when there is more than one.

**TRD-082** `render.Markdown` takes an outline as an option; without one it
uses `full`. A second built-in, `summary`, is a one-page plan: summary,
objectives, hypotheses and assumptions, included scope, client
responsibilities and contacts, deliverables (titles and due dates), the
phases and milestones as one schedule table, and signatures. The CLI takes
`--outline full|summary|<file>`. [PRD-FR-605]

**TRD-083** The content digest never depends on the outline. The footer and
the information table name the outline when it is not `full`.
[PRD-FR-607]

**TRD-084** `lint.Outline` reports an error for an unknown block or filter
and for a duplicate heading. Rendering an agreement in an outline warns when
a section's blocks are all empty for that agreement.

**TRD-085** `crosswalk.Build(agreement, outline)` returns, for each outline
section in order: its heading, the blocks it draws on, the section numbers
and titles those blocks have in the `full` layout, and whether they have
content. It also returns the blocks of the agreement with content that the
outline does not use. `cespec crosswalk` prints it as a Markdown table;
`render` and `export` take `--crosswalk <outline>` to append it to the
document as an appendix, "Correspondence with <outline title>". [PRD-FR-606]

**TRD-086** The repository ships only `full`, `summary`, and fictional
example outlines. An outline that follows another organization's form lives
in the user's repository. Outlines hold headings, block references, and the
user's own notes; the format has no field for imported template text.

