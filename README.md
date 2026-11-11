# Consulting Engagement Spec

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/ProductBuildersHQ/consulting-engagement-spec/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/ProductBuildersHQ/consulting-engagement-spec/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/ProductBuildersHQ/consulting-engagement-spec/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/ProductBuildersHQ/consulting-engagement-spec/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/ProductBuildersHQ/consulting-engagement-spec/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/ProductBuildersHQ/consulting-engagement-spec/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/ProductBuildersHQ/consulting-engagement-spec
 [docs-godoc-url]: https://pkg.go.dev/github.com/ProductBuildersHQ/consulting-engagement-spec
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/ProductBuildersHQ/consulting-engagement-spec/blob/main/LICENSE

**Consulting Engagement Spec** is a structured format from
[ProductBuildersHQ](https://github.com/ProductBuildersHQ) for consulting
agreements: proposals and statements of work. You write the agreement once,
in YAML or JSON. Tools check it, and render it to Markdown, DOCX, or a PDF the
parties sign.

It is a Go-first data model with a generated JSON Schema, a Go SDK, and the
`cespec` CLI.

## Why

Most proposals and statements of work are Word files. Nothing checks that the
scope says what is **out** of scope, that the client's responsibilities are
written down, or that every deliverable says how it will be accepted. Those
gaps cause most consulting disputes. And the PDF a client signs usually can't
be traced back to an exact source.

An agreement in this format has stable IDs for every deliverable, milestone,
responsibility, and exclusion, so it can be linted, compared between
versions, and later fed into a project plan.

## An agreement

```yaml
schemaVersion: "0.1"
kind: proposal
id: riverbend-impact-measurement
version: "0.1"
status: draft
title: Impact Measurement Work Plan
date: 2026-01-12
period:
  start: 2026-02-02
  end: 2026-06-30
parties:
  - id: northfield
    role: consultant
    name: Northfield Pro Bono Consultants
  - id: riverbend
    role: client
    name: Riverbend Food Network
people:
  - id: alex-lee
    name: Alex Lee
    party: northfield
    role: Project Leader
  - id: sam-ortiz
    name: Sam Ortiz
    party: riverbend
    title: Executive Director
background: |
  Riverbend reports outputs to funders but cannot yet show outcomes.
objectives:
  - id: O1
    title: Agree on a small set of outcome metrics
scope:
  included:
    - id: S1
      title: Metric framework and data collection plan
  excluded:
    - id: X1
      title: Building or hosting custom software
deliverables:
  - id: D1
    title: Metric framework
    due: 2026-03-27
    objectives: [O1]
    acceptanceCriteria:
      - Approved by the Executive Director
responsibilities:
  - id: R1
    party: riverbend
    description: Give read access to the volunteer database in week 1.
commercials:
  model: pro_bono
signatures:
  - party: northfield
    signer: alex-lee
  - party: riverbend
    signer: sam-ortiz
```

Quote `version`, `schemaVersion`, and fee `amount` values in YAML; they are
strings. See [`examples/`](examples/) for complete agreements.

## Quick start

```bash
go install github.com/ProductBuildersHQ/consulting-engagement-spec/cmd/cespec@latest

cespec validate agreement.yaml             # check against the JSON Schema
cespec lint agreement.yaml                 # references, dates, and gaps (--strict fails on warnings)
cespec fmt agreement.yaml                  # rewrite in canonical form (--check to report only)
cespec render agreement.yaml -o agreement.md                 # Pandoc Markdown
cespec render agreement.yaml --flavor gfm -o agreement.md    # GitHub Markdown
cespec export agreement.yaml -o agreement.docx               # needs pandoc
cespec export agreement.yaml -o agreement.pdf                # needs pandoc and lualatex
cespec digest agreement.yaml               # the content digest printed on the document
cespec schema agreement                    # print the embedded JSON Schema
```

`render` and `export` refuse to run while lint reports errors. Export pins the
timestamps Pandoc and LaTeX embed to the agreement's date, so the same input
produces the same file.

## Signing

The rendered document ends with a signature page: one block per signer, and
the agreement's **content digest**, a SHA-256 of its terms. The digest
ignores the status and the signing records, so it does not change when the
agreement is signed, but any change to the terms changes it. Anyone can run
`cespec digest` on the source and compare it with the signed copy.

Today the signature block works for wet ink, PDF fill-and-sign tools, and
e-signature services where the sender places the fields. Real PDF signature
fields, a frozen-PDF manifest, and PAdES signing through an external service
are planned (see the [roadmap](docs/specs/initiatives/INIT-CONSULTINGENGAGEMENTSPEC-001/ROADMAP.md)).

This format carries the terms the parties write. It is not legal advice;
have an attorney review terms and the signing process before relying on them.

## Specifications

- [PRD](docs/specs/initiatives/INIT-CONSULTINGENGAGEMENTSPEC-001/PRD.md) — problem, goals, requirements
- [TRD](docs/specs/initiatives/INIT-CONSULTINGENGAGEMENTSPEC-001/TRD.md) — data model, digest, lint rules, rendering, signing
- [PLAN](docs/specs/initiatives/INIT-CONSULTINGENGAGEMENTSPEC-001/PLAN.md) — sequencing and first adopter
- [ROADMAP](docs/specs/initiatives/INIT-CONSULTINGENGAGEMENTSPEC-001/ROADMAP.md) — phases and items

## Releases

- [CHANGELOG](CHANGELOG.md)
- [v0.1.0 release notes](docs/releases/v0.1.0.md)

## License

MIT
