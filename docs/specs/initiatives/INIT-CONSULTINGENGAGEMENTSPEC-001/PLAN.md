# PLAN — Consulting Engagement Specification

**Initiative:** `INIT-CONSULTINGENGAGEMENTSPEC-001`
**Repository:** `github.com/ProductBuildersHQ/consulting-engagement-spec`
**Workflow:** pbhq-lite
**Status:** Draft

## Approach

Build from a real engagement rather than an abstract schema. Phase 1 ships
the smallest thing that lets a team send a client a professional, checkable
agreement and get it signed. Each later phase is driven by what the first
adopter actually needs next.

## First adopter

The first adopter is a volunteer consulting team working pro bono with a
nonprofit client on impact measurement. The
team needs a work plan the client agrees to: background, objectives, scope and
exclusions, deliverables, milestones, each side's responsibilities, and sign
off by both parties.

Its agreement lives in the adopter's own repository, not here. **Examples in
this repository are fictional**; a real client's agreement, people, or terms
are never committed to it. The adopter's agreement is the acceptance test for
Phase 1: it must lint with zero errors and render to a DOCX and PDF the
client accepts without hand edits.

## Sequence

### Phase 1 — Agreement model, schema, and Markdown/Pandoc export

1. Scaffold the module, CI, and lint configuration.
2. Define the Agreement types and enumerations.
3. Generate and embed the JSON Schema; lint it with schemakit.
4. Load YAML and JSON strictly; canonical form; content digest.
5. Lint rules for references, dates, lifecycle, and completeness.
6. Deterministic Markdown rendering in Pandoc and GitHub flavors.
7. The `cespec` CLI, including Pandoc export to DOCX and PDF.
8. Fictional examples, golden files, and tests.
9. Adopter check: render the first real agreement and fix what it reveals.

Exit: the adopter's agreement is signed using the printed signature block.

### Phase 2 — Signable and signed PDFs

PDF signature fields, a freeze manifest with the hash of the frozen PDF,
importing signing records, PAdES signing through an external service, and a
verify command. Start only once Phase 1 is in use, and pick the signing
service with the adopter.

### Phase 3 — Lifecycle documents

Change orders, acceptance records, closure reports, and a diff between two
agreement versions. Each is a separate document kind that references an
agreement by ID and version.

### Phase 4 — Operations and integration

Turn a signed agreement into planning data (milestones and deliverables into
a PRISM roadmap; objectives linked to impact-model outcomes), an MCP server
for agents, and tracking in visionstudio.

## Risks

| Risk | Mitigation |
|---|---|
| The model is shaped too closely to one engagement | Keep fields generic; add a paid, fixed-fee fictional example alongside the pro bono one |
| PDF output differs across TeX installations | Pin timestamps; document the toolchain; hash the frozen PDF in Phase 2 rather than relying on cross-machine byte equality |
| Users treat the format as legal advice | State in README and PRD that terms are the parties' own and need legal review |
| Signing scope creeps into an e-signature platform | Phase 2 integrates an external service; identity and audit stay out of scope |
