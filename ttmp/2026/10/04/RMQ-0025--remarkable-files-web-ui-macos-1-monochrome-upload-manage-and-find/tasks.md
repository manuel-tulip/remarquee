# Tasks

## Documentation

- [x] Reconnaissance of existing cloud/upload/web layers
- [x] Create ticket and vocabulary topics
- [x] Write intern-facing design & implementation guide
- [x] Write implementation diary
- [x] Apply "no classic Macintosh bitmap typeface / modern fonts only" constraint
- [x] Resolve delivery shape as `remarquee serve` subcommand (not a separate binary)
- [x] Relate key source files to the design doc
- [x] Upload the design guide to reMarkable

## Implementation — DONE

### Phase 0 — Spike the service
- [x] Add `pkg/rmfiles/entries.go` with `Entry` + path building
- [x] Dump tree and compare with `remarquee cloud ls` (live: 13661 docs)

### Phase 1 — Service + read endpoints
- [x] Implement `pkg/rmfiles.Service` (single long-lived `ApiCtx`, mutex, reindex)
- [x] `GET /api/status`, `/api/files`, `/api/files/tree`, `/api/search`
- [x] Unit tests with a fake `ApiCtx`

### Phase 2 — Frontend shell + macOS-1 primitives
- [x] Static frontend at `cmd/remarquee/cmds/serve/frontend` (no build step)
- [x] `styles.css` tokens, file list, breadcrumbs, search, detail pane, status banner

### Phase 3 — Upload jobs
- [x] `POST /api/uploads` multipart + job store + bounded workers
- [x] `.md` → PDF conversion, `.pdf`/`.epub` direct; job polling UI
- [x] Verified live: create folder, md upload job `done`, recursive delete

### Phase 4 — Manage operations
- [x] Create folder, rename, move, delete (typed confirm), download
- [x] Guardrails: move-into-self rejection, conflict detection

### Phase 5 — Packaging + polish
- [x] `serve` command, `embed.go`, registered in `cmd/remarquee/main.go`
- [x] Graceful shutdown, `--addr` flag, `127.0.0.1` default, `--dev` cache toggle

### Phase 6 — Validation
- [x] `go test ./pkg/rmfiles/... ./cmd/remarquee/cmds/serve/...`
- [x] Live smoke test (tmux + curl) with cleanup

## Follow-ups (not done)

- [ ] Wire `rmcloud.WithAuthRetry` for mutating operations (signature requires a small service change)
- [ ] Same-origin/CSRF enforcement on mutating endpoints; optional read-only mode
- [ ] Opt-in integration test guarded by an env var
- [ ] Revisit `WriteTimeout: 0` and `Download` holding the mutex across I/O
