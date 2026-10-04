---
Title: 'Remarkable files web UI: macOS 1 monochrome upload, manage, and find'
Ticket: RMQ-0025
Status: complete
Topics:
    - remarkable
    - cloud
    - upload
    - web
    - frontend
    - ui
    - react
    - go
    - remarquee
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Design and implementation ticket for a local, macOS-1-styled web UI that uploads, manages, and quickly finds reMarkable cloud files by reusing remarquee's rmapi/rmcloud/mdpdf stack, delivered as the `remarquee serve` subcommand. Deliverable is an intern-facing guide; implementation is planned in phases."
LastUpdated: 2026-10-04T07:22:39.678261-04:00
WhatFor: "Explaining and implementing the reMarkable files web UI (Go HTTP API + embedded React SPA)."
WhenToUse: "When starting or reviewing implementation of the files web UI, or debugging the Go API / React frontend boundary."
---

# Remarkable files web UI: macOS 1 monochrome upload, manage, and find

## Overview

This ticket delivers a complete, intern-facing design and implementation guide
for a small local web application that lets a user **upload, manage, and quickly
find** their reMarkable cloud files. The UI is styled after classic Macintosh
System 1 (1984): 1-bit black-on-white, square corners, hairline rules, dithered
separators, **no window chrome and no menu bar**, and an accent palette used
**only as text highlights**. Fonts are modern — the classic Macintosh bitmap
typeface is explicitly not used.

Architecturally the app reuses almost all existing machinery: `pkg/rmcloud` for
auth/tree/download, `pkg/mdpdf` for Markdown→PDF, the `remarquee cloud …`
commands as the semantic reference, and `cmd/remarquee-ui` + the Dagger/pnpm
build helper as the Go-server-plus-embedded-SPA packaging precedent. The one
genuine engineering task is extracting a **single-context, mutex-serialized
service** (`pkg/rmfiles`) because `rmapi`'s `ApiCtx` is stateful and not safe for
concurrent mutation.

**Status: complete.** Implemented as `pkg/rmfiles` + `cmd/remarquee/cmds/serve`
(the `remarquee serve` subcommand) with an embedded no-build static frontend.
Validated against the live cloud; see the diary and tasks.md.

## Key Links

- [Design & implementation guide](design-doc/01-remarkable-files-web-ui-design-and-implementation-guide-for-a-new-intern.md)
- [Implementation diary](reference/01-implementation-diary.md)
- [Task list](tasks.md)
- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active** — design/guide complete, implementation planned
(Phases 0–6 in the design doc and tasks.md).

## Topics

- remarkable
- cloud
- upload
- web
- frontend
- ui
- react
- go
- remarquee

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
