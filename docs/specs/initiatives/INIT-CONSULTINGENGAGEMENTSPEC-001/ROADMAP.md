# ROADMAP — Consulting Engagement Specification — structured agreements rendered to DOCX and signable PDF

**Initiative:** `INIT-CONSULTINGENGAGEMENTSPEC-001`
**Repository:** `github.com/ProductBuildersHQ/consulting-engagement-spec`
**Status:** In progress — 0 of 5 phases completed

> RMI IDs are stable and permanent. Commits implementing an item carry its own ID as the trailer, `Refs: <RMI-ID>`. Phase status is derived from member RMIs — a phase is complete only when all its RMIs are complete. Requirements in each `Satisfies` line are defined in PRD.md and TRD.md.

## Phase 1 — Agreement Model, Schema, and Markdown/Pandoc Export

**Theme:** One structured Agreement document, checked beyond its schema, rendered deterministically to Markdown, DOCX, and PDF with a printed signature block
**Status:** In progress — 8 of 9 items completed

- [x] `RMI-CONSULTINGENGAGEMENTSPEC-001` Repository scaffold: go.mod, CI workflows (go-ci, go-lint, go-sast), lint config, README, CLAUDE.md, LICENSE
  - Type: enabler · Status: completed
  - Satisfies: — (enabler)
  - Acceptance: `go build ./...` and `golangci-lint run` pass; CI workflows green
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-002` `pkg/engagement` Agreement types, enumerations with `All…`/`Valid` helpers, and lookups
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-001`
  - Satisfies: PRD-FR-101…108, TRD-002…005
  - Acceptance: every field in TRD §2 exists with its JSON name; enum helpers tested
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-003` Schema pipeline: `pkg/schemagen`, `cmd/generate-schema`, embedded `schema/` with Names/Get/Compiled/Validate
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-002`
  - Satisfies: PRD-FR-201, TRD-010…012
  - Acceptance: `schemakit lint --property-case camelCase` reports no issues; regeneration leaves no diff
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-004` `pkg/engagementio` YAML/JSON loading with limits and strict decoding, canonical form, and content digest
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-003`
  - Satisfies: PRD-FR-202, PRD-FR-203, PRD-FR-403, TRD-020…023, TRD-030
  - Acceptance: YAML and JSON of the same agreement give the same digest; canonical output is idempotent; digest ignores status and signing records
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-005` `pkg/lint` rules for IDs, references, dates, commercials, lifecycle, and completeness
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-002`
  - Satisfies: PRD-FR-301, PRD-FR-302, TRD-040
  - Acceptance: every rule in TRD §6 has a triggering test case; output order is deterministic
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-006` `pkg/render` deterministic Markdown in Pandoc and GitHub flavors, with information table, numbered sections, and signature blocks
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-004`
  - Satisfies: PRD-FR-401, PRD-FR-403, PRD-FR-404, TRD-050…053
  - Acceptance: golden files match byte for byte; absent sections are omitted and numbering stays contiguous
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-007` `pkg/pandoc` command builder and the `cespec` CLI (validate, lint, fmt, digest, render, export, schema)
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-005`, `RMI-CONSULTINGENGAGEMENTSPEC-006`
  - Satisfies: PRD-FR-402, PRD-FR-501, TRD-001, TRD-060…062
  - Acceptance: `cespec export` writes DOCX and PDF; two exports of the same input produce the same DOCX and PDF bytes
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-008` Fictional pro bono example, golden files, and tests across packages
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-007`
  - Satisfies: PRD §7
  - Acceptance: every example validates and lints with zero errors; rendered goldens are checked in and tested
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-009` Adopter check: render the first real engagement agreement and fold back what it reveals
  - Type: validation · Status: in_progress
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-008`
  - Satisfies: PRD §7
  - Acceptance: the adopter's agreement lints with zero errors and the client accepts the DOCX or PDF without hand edits

## Phase 2 — Signable and Signed PDFs

**Theme:** A PDF readers can sign, tied to a frozen, hashed version of the agreement and verifiable afterwards
**Status:** Planned — 0 of 5 items completed

- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-010` PDF export with AcroForm signature fields, one per signature block
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-007`
  - Satisfies: PRD-FR-405, TRD-071
  - Acceptance: Acrobat and one other reader show an unsigned signature field for each signer
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-011` Freeze: a manifest with agreement ID, version, content digest, toolchain versions, and the SHA-256 of the frozen PDF
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-010`
  - Satisfies: PRD-FR-405, TRD-071
  - Acceptance: freezing twice with the same toolchain gives the same PDF hash; the manifest validates against a generated schema
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-012` Record signing: import signed dates and methods into the agreement and mark it agreed
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-011`
  - Satisfies: PRD-FR-107, TRD-030
  - Acceptance: the content digest is unchanged by the import; lint reports no `agreed-unsigned` once all have signed
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-013` PAdES signing through an external signing service
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-011`
  - Satisfies: PRD-FR-405, TRD-071, TRD-072
  - Acceptance: a frozen PDF is sent, signed by each signer, and returned with a valid PAdES signature
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-014` `cespec verify`: check a signed PDF against its manifest and source agreement
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-012`, `RMI-CONSULTINGENGAGEMENTSPEC-013`
  - Satisfies: PRD-FR-405, TRD-071
  - Acceptance: verify passes for an untouched signed PDF and fails when the source agreement or frozen PDF changes

## Phase 3 — Lifecycle Documents

**Theme:** Documents for the stages after agreement, each referencing an agreement by ID and version
**Status:** Planned — 0 of 4 items completed

- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-015` Change order: added, removed, and changed scope, deliverables, dates, and fees against a signed agreement
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-008`
  - Satisfies: PRD §4
  - Acceptance: a change order validates, lints its references against the agreement, and renders with signature blocks
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-016` Acceptance record: formal acceptance of deliverables, with exceptions
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-008`
  - Satisfies: PRD §4
  - Acceptance: lint reports deliverables not yet accepted
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-017` Closure report: handoff, outcomes against success criteria, and completion sign-off
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-016`
  - Satisfies: PRD §4
  - Acceptance: renders from an agreement plus its acceptance records
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-018` `cespec diff`: element-level differences between two agreement versions
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-004`
  - Satisfies: PRD-FR-104
  - Acceptance: reports added, removed, and changed elements by ID, in a deterministic order

## Phase 4 — Operations and Integration

**Theme:** Signed agreements feed planning tools and agents without retyping
**Status:** Planned — 0 of 4 items completed

- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-019` Export milestones, deliverables, and responsibilities as a PRISM roadmap
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-008`
  - Satisfies: PRD G-5
  - Acceptance: the export loads in the PRISM roadmap tools with IDs preserved
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-020` Link objectives and success criteria to impact-model outcomes and indicators by reference
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-008`
  - Satisfies: PRD G-5
  - Acceptance: references resolve against an impact model, and lint reports broken ones
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-021` MCP server over the service layer: validate, lint, render, digest
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-007`
  - Satisfies: PRD-FR-501
  - Acceptance: an agent can draft, lint, and render an agreement through MCP alone
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-022` Track engagements in visionstudio: agreement status, milestones, and acceptance
  - Type: capability · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-016`
  - Satisfies: PRD G-5
  - Acceptance: an engagement's agreement and acceptance state appear in visionstudio

## Phase 5 — Alignment with Other Forms

**Theme:** One agreement, laid out in any organization's sections and order, with a crosswalk that shows it is a superset
**Status:** In progress — 7 of 8 items completed

- [x] `RMI-CONSULTINGENGAGEMENTSPEC-023` Hypotheses: types, schema, lint, and the Hypotheses section
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-008`
  - Satisfies: PRD-FR-601, TRD-006, TRD-052
  - Acceptance: a hypothesis with a test, deliverables, and a milestone renders under Hypotheses; unknown references are errors and an untested hypothesis is a warning
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-024` Milestone kinds and closeout: evaluations and reports owed to a party
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-008`
  - Satisfies: PRD-FR-602, PRD-FR-603
  - Acceptance: kinds print with milestones; the Closeout section lists evaluations and reports; `no-kickoff` and `no-closeout` warn only when some milestone has a kind
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-025` Points of contact and signer capacity
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-008`
  - Satisfies: PRD-FR-604
  - Acceptance: contacts print under Working together; a capacity prints in place of the signer's title; several signature blocks per party render and lint cleanly
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-026` Outline document kind: types, generated schema, loading, and lint
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-023`, `RMI-CONSULTINGENGAGEMENTSPEC-024`, `RMI-CONSULTINGENGAGEMENTSPEC-025`
  - Satisfies: PRD-FR-605, TRD-080, TRD-081, TRD-084
  - Acceptance: an outline loads from YAML or JSON, validates against its embedded schema, and lint reports unknown blocks, unknown filters, and duplicate headings
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-027` Render in an outline: built-in `full` and `summary`, block filters, and month dates
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-026`
  - Satisfies: PRD-FR-605, PRD-FR-607, TRD-082, TRD-083
  - Acceptance: `render` and `export` take `--outline`; `full` reproduces the default layout byte for byte; the digest is the same in every outline
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-028` Crosswalk: `cespec crosswalk` and an appendix in the rendered agreement
  - Type: capability · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-027`
  - Satisfies: PRD-FR-606, TRD-085
  - Acceptance: the table lists each outline section with the full-layout sections that answer it and whether they have content, then the blocks the outline leaves out; deterministic output
- [x] `RMI-CONSULTINGENGAGEMENTSPEC-029` Fictional example outline, golden files, and a guide to aligning with an organization's form
  - Type: enabler · Status: completed
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-028`
  - Satisfies: PRD G-6, TRD-086
  - Acceptance: the example renders in its outline and crosswalk with zero lint findings; the guide explains keeping an organization's outline in the user's own repository
- [ ] `RMI-CONSULTINGENGAGEMENTSPEC-030` Adopter check: render the first real agreement in its program's work plan layout from a private outline
  - Type: enabler · Status: planned
  - Depends on: `RMI-CONSULTINGENGAGEMENTSPEC-029`
  - Satisfies: PRD G-6
  - Acceptance: the adopter's agreement renders in its program's layout, and the crosswalk shows every section of that layout answered
