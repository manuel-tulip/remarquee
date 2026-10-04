---
Title: Implementation diary
Ticket: RMQ-0025
Status: draft
Topics:
    - remarkable
    - cloud
    - upload
    - web
    - frontend
    - ui
    - remarquee
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological diary for RMQ-0025: reconnaissance and the intern-facing design/implementation guide for the reMarkable files web UI."
LastUpdated: 2026-10-04T00:00:00-04:00
---

# Implementation diary — RMQ-0025

## Step 1: Reconnaissance, ticket scaffolding, and the intern design guide

I started by mapping the repository's existing cloud and upload machinery so the
design could reuse it rather than invent parallel logic. The key discovery was
that the repo already contains both halves of the answer: a complete
rmapi-backed cloud/upload CLI layer (`cmd/remarquee/cmds/cloud`,
`cmd/remarquee/cmds/upload`, `pkg/rmcloud`, `pkg/mdpdf`) and a working Go +
embedded-React web-app precedent (`cmd/remarquee-ui` with a Dagger/pnpm build
helper). I then created the ticket, added the design guide and this diary,
promoted four new vocabulary topics, and left the relative file links to be
added once the doc was in place.

### Prompt Context
**User prompt (verbatim):** "Create a new docmgr ticket to add a simple web UI in a retro monochrome macos1 style (modern fonts, no window chrome, no menu bar, just the overall look, simple color highlights for font color) to allow me to upload and manage and quickly find my remarkable files.

Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet points and pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and upload to remarkable."

**User follow-up (verbatim):** "no chicago"

**Assistant interpretation:** Produce a design/implementation ticket (not working
code) for a local macOS-1-styled web UI that uploads/manages/searches reMarkable
cloud files, written for an intern, stored in the ticket and delivered to the
reMarkable cloud. The follow-up clarifies that the classic Macintosh bitmap
typeface must not be used; modern fonts only.

**Inferred user intent:** A build-ready blueprint the user can hand to a new
engineer, plus a readable PDF on their tablet, with an unambiguous visual brief
(monochrome, square, no window chrome/menu bar, color only as text highlights,
modern fonts).

**Commit (code):** N/A — documentation-only ticket, not yet committed.

### What I did

- Ran `docmgr status --summary-only` (root `ttmp`, 36 tickets, 105 docs) and
  `docmgr vocab list` to learn the allowed topics/doc-types.
- Added vocabulary topics: `web`, `frontend`, `ui`, `react`.
- Created ticket `RMQ-0025` titled "Remarkable files web UI: macOS 1 monochrome
  upload, manage, and find" at
  `ttmp/2026/10/04/RMQ-0025--remarkable-files-web-ui-macos-1-monochrome-upload-manage-and-find/`.
- Added a `design-doc` ("…design and implementation guide for a new intern")
  and a `reference` diary.
- Read and cited: `pkg/rmcloud/auth.go` (`CreateApiCtx`, `forceSchemaV4`,
  `WithAuthRetry`), `pkg/rmcloud/dirs.go` (`MkdirAll`),
  `cmd/remarquee/cmds/cloud/{ls,find,search,put,get,rm,mv}.go`,
  `cmd/remarquee/cmds/upload/md.go`, `pkg/mdpdf/pandoc.go`,
  `cmd/remarquee-ui/{main.go,embed.go,Makefile}`,
  `cmd/build-remarquee-ui-web/main.go`,
  `cmd/remarquee-ui/frontend/{package.json,vite.config.ts}`, and the pinned
  `rmapi` fork's `api/api.go`, `filetree/filetree.go`, `model/*.go`,
  `api/sync15/common.go`, `config/config.go`.
- Verified the dependency pin: `go.mod` has
  `replace github.com/juruen/rmapi => github.com/FNStudios-NI/rmapi v0.0.0-20260817154736-f295d5466978`.
- Wrote the guide (~15 sections incl. appendices) with prose, bullets,
  pseudocode, ASCII diagrams, an HTTP API table, a design-system token block,
  a phased plan, a test strategy, and file/symbol references.
- Applied the "no chicago" follow-up by removing all uses of the bitmap
  typeface from the requirements, vocabulary, and design-system sections and
  replacing them with an explicit "modern fonts only; classic bitmap typeface
  not used" constraint.

### Why

- The repo's prior ticket `RMQ-RMDOC-WEB-001`/`cmd/remarquee-ui` is the exact
  packaging precedent for a Go server with an embedded SPA; reusing it avoids
  inventing a second frontend toolchain.
- All cloud semantics (upload overwrite/annotation loss, delete confirmation,
  template filtering, path synthesis) already live in the CLI; the guide points
  the intern at those exact files so the UI reproduces behavior rather than
  guessing it.
- Considered implementing code, but the user explicitly asked for an
  analysis/design/implementation guide to be stored and uploaded, so the
  deliverable is documentation.

### What worked

- `docmgr ticket create-ticket` and `docmgr doc add` produced the standard
  workspace quickly.
- Reading the pinned fork source directly (module cache path with the
  `FNStudios-NI` escape) confirmed the `ApiCtx` interface, `model.Node` fields,
  and the `tree.cache` location used in the guide.

### What didn't work

- The first attempt to locate the dependency under
  `$(go env GOMODCACHE)/github.com/juruen/rmapi@v0.0.25` failed (module not
  found) because of the `replace` directive. Recovery: read `go.mod` and used
  the `FNStudios-NI/rmapi@v0.0.0-20260817154736-f295d5466978` path.

### What I learned

- `ApiCtx` is stateful and non-thread-safe; `rmcloud.CreateApiCtx` eagerly
  loads/mirrors/saves a cached tree and retries auth up to 3×. This is the
  central constraint for a web server, so it became §4.4 of the guide.
- The cloud has no stable user path; paths are synthesized by walking parent
  pointers, so document **IDs** must be the UI's primary key.

### What was tricky to build

- Deciding the service boundary: command-layer logic is embedded in Cobra `Run`
  functions, so the UI cannot call it without refactoring. The guide proposes
  `pkg/rmfiles` with a single long-lived `ApiCtx` and a mutex, conversion
  outside the lock, and `SyncComplete()` + reindex after mutations. Sequencing
  (convert → lock → upload → sync → reindex) is the part most likely to be got
  wrong.

### What warrants a second pair of eyes

- Concurrency: the single-context + mutex design and the split between
  conversion and the tree critical section.
- Security: localhost binding, CSRF/same-origin checks on mutating endpoints,
  path-traversal guards on download, and never serializing tokens.
- Destructive semantics: delete confirmation and overwrite (annotation loss)
  must mirror the CLI's safety gates.

### What should be done in the future

- Add `RelatedFiles` to the design doc and ticket index with absolute paths.
- Implement Phases 0–6; refactor `pkg/rmfiles` + the shared web build helper.
- Upload the guide to reMarkable.

### Code review instructions

- Start with the design doc Part 2 (§4.1–4.4) and Part 4 (§6.2–6.6).
- Cross-check claims against `pkg/rmcloud/auth.go`, `cmd/remarquee-ui/embed.go`,
  and the pinned `rmapi/api/api.go`.
- No code was changed in this step; validation is the doc review itself.

### Technical details

- Module: `github.com/go-go-golems/remarquee`, Go 1.26.3.
- Pinned rmapi fork: `github.com/FNStudios-NI/rmapi
  v0.0.0-20260817154736-f295d5466978` (via `go.mod` `replace`).
- Tokens: `~/.rmapi` (override `RMAPI_CONFIG`; else XDG
  `rmapi/rmapi.conf`). Tree cache: `os.UserCacheDir()/rmapi/tree.cache`,
  `CacheVersion = 3`.
- `ApiCtx`: `Filetree`, `FetchDocument`, `CreateDir`, `UploadDocument`,
  `ReplaceDocumentFile`, `MoveEntry`, `DeleteEntry`, `SyncComplete`, `Nuke`,
  `Refresh`.
- Existing web precedent: `net/http.ServeMux` + `//go:embed frontend/dist` +
  Vite/React 19/Redux Toolkit + Dagger/pnpm helper
  (`cmd/build-remarquee-ui-web`).

## Step 2: Resolve delivery shape as `remarquee serve`, not a separate binary

The user clarified two visual/packaging constraints after the first draft: the
classic Macintosh bitmap typeface must not be used (recorded in Step 1), and the
web UI must be delivered **as a CLI subcommand (`remarquee serve` or similar),
not a separate binary**. I updated the design guide accordingly so the proposed
architecture, packaging, and open questions all reflect a subcommand of the
existing `remarquee` binary.

### Prompt Context
**User prompt (verbatim):** "do it as `remarquee serve` or so, not separate binary"
**Assistant interpretation:** Change the planned command/packaging shape from a
standalone `cmd/remarquee-web` binary to a `serve` subcommand registered in the
existing CLI.
**Inferred user intent:** One binary, one install path, discoverable help, and
the embedded SPA shipped with the existing release — do not add a second
executable.

### What I did

- Updated the executive summary and the §3.2 architecture diagram to show
  `remarquee serve (cmd/remarquee/cmds/serve)` as the HTTP server.
- Replaced the "standalone binary" decision in §4.6 with a resolved decision to
  add a `serve` subcommand and register it in `cmd/remarquee/main.go`.
- Rewrote the §6.2 proposed code layout to
  `cmd/remarquee/cmds/serve/{command.go,server.go,embed.go,gen.go,frontend/}`
  plus `pkg/rmfiles/`, and documented the registration and lifecycle
  (`RunE` builds the service + `http.Server`, blocks until context cancellation;
  SIGINT is already propagated by `executeInterruptible`).
- Updated §6.11 dev/prod, the Phase 5 exit criteria, the test command paths, the
  manual smoke command (`go run ./cmd/remarquee serve --dev`), the "done"
  definition, and open question #1 (now RESOLVED).
- Added Appendix A rows for `cmd/remarquee/main.go` (registration) and clarified
  the frontend location.

### Why

- The repo already has one flag-driven `main` command; a subcommand gives users
  `remarquee serve --help` and avoids a second GoReleaser/Homebrew artifact.
- `executeInterruptible` in `cmd/remarquee/main.go` already handles SIGINT for
  long-running verbs, which fits a server command.

### What worked

- The changes were localized to the guide; no code exists yet, so there was
  nothing to compile.

### What didn't work

- No failures observed.

### What I learned

- The existing `remarquee-ui` binary can stay untouched; only the *new* UI is a
  subcommand. The two can share the Dagger/pnpm build helper via
  `RMQ_WEB_DIR`/`RMQ_WEB_DIST`.

### What warrants a second pair of eyes

- Whether embedding the SPA inside `cmd/remarquee` measurably affects CLI build
  time/size and whether `go:embed frontend/dist` in a subpackage is acceptable
  for the release pipeline.

### What should be done in the future

- Implement `cmd/remarquee/cmds/serve` and register it; run the upload again once
  the guide is final.

### Code review instructions

- Confirm no `cmd/remarquee-web` references remain in the guide
  (`grep -n "remarquee-web" design-doc/*.md` should be empty except history).
- Spot-check `cmd/remarquee/main.go` registration pattern.

### Technical details

- Proposed files: `cmd/remarquee/cmds/serve/{command.go,server.go,embed.go,gen.go,frontend/}`.
- Registration: `rootCmd.AddCommand(serve_cmd.NewServeCommand())` in
  `cmd/remarquee/main.go`.
- Invocation: `remarquee serve [--addr 127.0.0.1:8080] [--dev] [--remote-dir /ai] [--workers N]`.
- Build: `go generate ./cmd/remarquee/cmds/serve` then `go build ./cmd/remarquee`.

## Step 3: Upload the guide to reMarkable

With the guide final (including both follow-up constraints), I uploaded the
design document to the reMarkable cloud as a single PDF using the remarquee
upload pipeline.

### Prompt Context
**User prompt (verbatim):** see Step 1 ("… upload to remarkable").
**Assistant interpretation:** Deliver the final design guide to the user's
reMarkable cloud.
**Inferred user intent:** Read the guide on the tablet.

### What I did

```bash
remarquee upload md \
  ttmp/2026/10/04/RMQ-0025--remarkable-files-web-ui-macos-1-monochrome-upload-manage-and-find/design-doc/01-remarkable-files-web-ui-design-and-implementation-guide-for-a-new-intern.md \
  --name "RMQ-0025 Remarkable Files Web UI - Intern Design Guide" \
  --remote-dir "/ai/2026/10/04/RMQ-0025" \
  --non-interactive
```

### What worked

```text
OK: uploaded RMQ-0025_Remarkable_Files_Web_UI_-_Intern_Design_Guide.pdf -> /ai/2026/10/04/RMQ-0025
Cloud tree synchronized. 0s | 1 HTTP responses
```

Delivery evidence: `OK: uploaded … -> /ai/2026/10/04/RMQ-0025`. This proves cloud
delivery, not physical device synchronization.

### What didn't work

- The first upload attempt was interrupted before completion; re-running the
  command after the guide was finalized succeeded cleanly.

### What warrants a second pair of eyes

- Verify the rendered PDF flows well on the tablet (wide ASCII diagrams and long
  identifier strings are the risk). A visual check on the device is the
  recommended follow-up.

### Technical details

- Document name: `RMQ-0025_Remarkable_Files_Web_UI_-_Intern_Design_Guide.pdf`
- Destination: `/ai/2026/10/04/RMQ-0025`
