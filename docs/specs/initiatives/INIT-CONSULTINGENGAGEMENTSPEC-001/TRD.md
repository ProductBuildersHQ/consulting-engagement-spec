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
| `scope` | yes | `summary`, `included` (≥1), `excluded` |
| `approach` | no | Method (Markdown) |
| `phases` | no | Stages of the work, with optional dates and goal |
| `deliverables` | yes, ≥1 | Outputs, with owner, phase, due date, format, objectives served, acceptance criteria |
| `milestones` | no | Dated checkpoints, with the decision taken and deliverables reviewed |
| `responsibilities` | yes, ≥1 | What each party commits to do, with optional due date |
| `governance` | no | `summary` and `meetings` (title, participants, frequency, duration) |
| `assumptions` | no | Conditions the plan depends on |
| `risks` | no | Risk, mitigation, owning party |
| `successCriteria` | no | Measures of whether the engagement succeeded |
| `acceptance` | no | Review `process` and `reviewDays` |
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
- **Signature:** `party`, `signer` (person ID), `signedDate`, `method`
  (`wet_ink`, `electronic`, `digital`). [PRD-FR-107]

**TRD-003** Element IDs of objectives, scope items, phases, deliverables,
milestones, responsibilities, assumptions, risks, criteria, terms, and open
questions match `^[A-Za-z0-9][A-Za-z0-9._-]*$`, so authors can use short
labels such as `D1` or `M2`. Party and person IDs are lowercase slugs.
[PRD-FR-104]

**TRD-004** All IDs share one namespace: parties, people, and every element
kind. A reference such as an open question's `owner` (a party or a person)
is therefore never ambiguous. [PRD-FR-104]

**TRD-005** Narrative fields are Markdown. In list items and table cells the
renderer collapses them to one line, so multi-paragraph text belongs in
`summary`, `background`, `approach`, `changeControl`, and `terms`.
[PRD-FR-108]

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
4. Scope (In scope; Out of scope)
5. Approach (with a Phases table)
6. Deliverables (title, phase, due date, non-consultant owner, description,
   format, "Accepted when")
7. Milestones
8. Responsibilities (grouped by party, in party order)
9. Team and contacts
10. Working together (governance and meetings)
11. Assumptions
12. Risks
13. Measures of success
14. Review and acceptance
15. Fees and expenses (with a total when there is more than one fee)
16. Changes to this agreement
17. Terms
18. Open questions
19. Signatures

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
