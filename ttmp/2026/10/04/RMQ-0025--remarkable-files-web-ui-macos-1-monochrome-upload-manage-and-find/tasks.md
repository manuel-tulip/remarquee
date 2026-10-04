# Tasks

## Documentation (this ticket)

- [x] Reconnaissance of existing cloud/upload/web layers
- [x] Create ticket and vocabulary topics
- [x] Write intern-facing design & implementation guide
- [x] Write implementation diary
- [x] Apply "no classic Macintosh bitmap typeface / modern fonts only" constraint
- [x] Relate key source files to the design doc
- [x] Upload the design guide to reMarkable

## Implementation (follow-up work planned in the guide)

### Phase 0 — Spike the service
- [ ] Add `pkg/rmfiles/entries.go` with `Entry` + path building
- [ ] Dump `WalkTree` as JSON and compare with `remarquee cloud ls --with-glaze-output --format json`

### Phase 1 — Service + read endpoints
- [ ] Implement `pkg/rmfiles.Service` (single long-lived `ApiCtx`, mutex, reindex)
- [ ] `GET /api/status`, `/api/files`, `/api/files/tree`, `/api/search`
- [ ] Unit tests with a fake `ApiCtx`

### Phase 2 — Frontend shell + macOS-1 primitives
- [ ] Scaffold `cmd/remarquee/cmds/serve/frontend` (Vite + React + TS)
- [ ] `tokens.css` / `base.css`, `FileList`, `Breadcrumbs`, `SearchBar`, `StatusBanner`
- [ ] Dev proxy `/api` → `:8080`

### Phase 3 — Upload jobs
- [ ] `POST /api/uploads` multipart + job store + bounded workers
- [ ] Single `.md` / `.pdf`, then directory + bundle
- [ ] `UploadDropzone` + job polling UI

### Phase 4 — Manage operations
- [ ] Create folder, rename, move, delete (typed confirm), download
- [ ] Guardrails: overwrite warning, move-into-self rejection

### Phase 5 — Packaging + polish
- [ ] Add `serve/{command.go,server.go,embed.go,gen.go}` and register `serve` in `cmd/remarquee/main.go`
- [ ] Reuse/parametrize the web build helper; `go build ./cmd/remarquee`
- [ ] Graceful shutdown, `--addr` flag, `127.0.0.1` default

### Phase 6 — Validation + docs
- [ ] `go test ./pkg/rmfiles/... ./cmd/remarquee/cmds/serve/...`
- [ ] Manual tmux smoke script
