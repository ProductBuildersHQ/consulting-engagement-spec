# CLAUDE.md — consulting-engagement-spec

Go-first consulting engagement specification: Go structs in `pkg/engagement`
are the source of truth; JSON Schemas under `schema/` are generated from them
and embedded. The `cespec` CLI (Cobra) is a thin layer over the library
packages. The first document kind is the Agreement (a proposal or statement
of work), rendered to Markdown and, through Pandoc, to DOCX and PDF.

## Specs first

Read before implementing:

- `docs/specs/initiatives/INIT-CONSULTINGENGAGEMENTSPEC-001/PRD.md` — what and why
- `docs/specs/initiatives/INIT-CONSULTINGENGAGEMENTSPEC-001/TRD.md` — data
  model, digest, lint rules, rendering, export, signing
- `docs/specs/initiatives/INIT-CONSULTINGENGAGEMENTSPEC-001/PLAN.md` — sequencing
- `docs/specs/initiatives/INIT-CONSULTINGENGAGEMENTSPEC-001/ROADMAP.md` — RMI ledger

## Conventions

- **Go-first schemas.** Never hand-edit `schema/*.schema.json`; regenerate
  with `go generate ./schema/...` (or `go run ./cmd/generate-schema -out schema`)
  and lint with `schemakit lint --property-case camelCase`.
- **JSON property names** are camelCase; enum values are snake_case. Enum
  values come from the `All…` functions in `pkg/engagement`.
- **Library-first.** Logic lives in `pkg/...`; `cmd/` and `internal/cli`
  stay thin. Library packages do no network I/O and start no processes;
  `pkg/pandoc` builds the Pandoc command and the CLI runs it.
- **Determinism.** Same inputs produce byte-identical outputs; never let map
  iteration order reach output. Pandoc export pins timestamps with
  `SOURCE_DATE_EPOCH`.
- **Content digest.** `engagementio.ContentDigest` ignores status and signing
  records by design. Do not add fields to it that change when a document is
  signed.
- **Examples are fictional.** Never commit a real client's agreement, people,
  or terms to this repository. Real agreements live in the adopter's own
  repository.
- **Not legal advice.** Do not add legal language presented as a standard
  template.

## VisionStudio

Roadmap items are tracked in VisionStudio under
`INIT-CONSULTINGENGAGEMENTSPEC-001`, whose home repository is
`github.com/ProductBuildersHQ/consulting-engagement-spec`. Claim work with
`visionstudio work claim <RMI-ID>` and carry the item's own ID as the trailer
on every implementing commit: `Refs: RMI-CONSULTINGENGAGEMENTSPEC-<NNN>`.

## Checks before committing

```bash
gofmt -l .
go build ./...
go test ./...
golangci-lint run
go generate ./schema/... && git diff --exit-code schema/
```
