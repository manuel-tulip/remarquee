# Changelog

## 2026-10-04

- Initial workspace created

## 2026-10-04

Created RMQ-0025 and wrote the intern-facing design/implementation guide for a macOS-1 monochrome web UI to upload, manage, and search reMarkable cloud files. Applied the 'modern fonts only, no classic bitmap typeface' constraint. Added web/frontend/ui/react vocabulary topics.

### Related Files

- /Users/manuel.odendahl/code/go-go-golems/remarquee/ttmp/2026/10/04/RMQ-0025--remarkable-files-web-ui-macos-1-monochrome-upload-manage-and-find/design-doc/01-remarkable-files-web-ui-design-and-implementation-guide-for-a-new-intern.md — Primary design and implementation guide
- /Users/manuel.odendahl/code/go-go-golems/remarquee/ttmp/2026/10/04/RMQ-0025--remarkable-files-web-ui-macos-1-monochrome-upload-manage-and-find/reference/01-implementation-diary.md — Reconnaissance diary and evidence

## 2026-10-04

Recorded two user constraints in the guide: modern fonts only (classic Macintosh bitmap typeface explicitly excluded) and delivery as a remarquee serve subcommand rather than a separate binary. Updated architecture, code layout, packaging, tests, and the resolved open question.

## 2026-10-04

Uploaded the finalized design guide to reMarkable: OK: uploaded RMQ-0025_Remarkable_Files_Web_UI_-_Intern_Design_Guide.pdf -> /ai/2026/10/04/RMQ-0025

## 2026-10-04

Implemented the feature: pkg/rmfiles (single-context, mutex-serialized service with search index, manage ops, and upload jobs) and the remarquee serve subcommand with an embedded no-build macOS-1 monochrome frontend. Registered serve in cmd/remarquee/main.go. Added unit + handler tests; validated end-to-end against the live cloud (13661 docs, create folder + md upload + recursive delete). Commits f62e278, 0a692bb, 137d9dc.

## 2026-10-04

Added folder uploads with preserved directory structure and browser drag-and-drop. Backend accepts folder-relative paths (sanitized) and creates intermediate remote folders; frontend stages files from a folder picker or drag-and-drop and sends parallel paths. Verified live (nested md+pdf upload and recursive delete). Commits 30b228e, bef63ad, 76997c4.
