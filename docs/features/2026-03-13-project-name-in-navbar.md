---
id: FEAT-003
date: 2026-03-13
status: draft
type: feature
research_level: light
plan: docs/plans/2026-03-13-project-name-in-navbar.md
yagni_verdict: build
tags: [ui, navigation, navbar, multi-instance]
---

# Project Name in Navigation Bar

> **Value:** Makes each `kr` instance immediately identifiable by showing the project folder name in the nav bar and browser tab title — eliminating the "which tab is which project" problem when running multiple instances.

## Problem

When running `kr` on multiple projects at the same time, every browser tab shows "kr" in both the page title and the nav bar brand. There is no visual cue to distinguish which instance is serving which project without looking at the URL or clicking around.

**Trigger:** Daily multi-project usage — the user opens `kr` for several projects simultaneously and loses track of which tab belongs to which project.
**Current workaround:** Check the URL path, look at folder names in the nav — both require effort and are error-prone. There is no workaround that gives a glance-level answer.

## YAGNI Assessment

**Verdict:** BUILD IT

This is a real, daily pain with a minimal, surgical fix. The project name is already available via `filepath.Base(rootPath)` — no new state, no new flags, no design decisions. The change is: one new field on three structs, three handler lines, two template references. Nothing to trim.

## Solution

### What we're building

1. **Project name in nav brand:** The `kr` text in `<a href="/" class="nav-brand">` becomes the name of the folder being served (e.g., `myapp-docs`).

2. **Project name in browser tab title:** The `<title>` tag changes from the hardcoded `kr` to the folder name, so each browser tab is uniquely titled.

### How it works

1. On startup, `Server` derives the project name once: `filepath.Base(rootPath)`
2. Each handler populates a `ProjectName` field in its template data struct
3. `layout.html` uses `{{.ProjectName}}` in both `<title>` and the nav brand link

### Visual concept

```
Before:
┌─────────────────────────────────────────────────┐
│  kr  │ Backlog │ Bugs │ Features │ Plans │ ...   │
└─────────────────────────────────────────────────┘
Browser tab: "kr"

After (running with --path /projects/myapp/docs):
┌─────────────────────────────────────────────────┐
│  docs  │ Backlog │ Bugs │ Features │ Plans │ ... │
└─────────────────────────────────────────────────┘
Browser tab: "docs"
```

## Boundaries

### Explicitly NOT building
- `--name` flag for a custom display name — the folder name is always descriptive enough for the stated use case
- Per-page title suffixes (e.g., "docs / Features") — simple project name solves the problem without adding complexity
- Favicon differentiation per project — out of scope

### Rabbit holes to avoid
- **Path normalization edge cases** — `filepath.Base` handles trailing slashes and relative paths correctly. Don't add custom logic.
- **Embedding project name at build time** — the name is runtime data from `--path`, not a build-time constant.

## Definition of Done

**The feature is complete when:**

1. Running `kr --path /any/path/to/projectname` shows `projectname` in the nav bar brand on all page types (backlog, folder, document)
2. The browser tab title shows `projectname` (not "kr") on all page types
3. The brand link still navigates to `/` (the Kanban board)

**Verification:**

Automated:
- [ ] `go test ./...` — all existing tests pass
- [ ] `go build -o kr .` — binary builds without errors

Manual:
- [ ] Run `kr --path ./docs` from the project root — nav brand shows "docs", browser tab shows "docs"
- [ ] Navigate to a folder — project name persists in nav and title
- [ ] Navigate to a document — project name persists in nav and title
- [ ] Run two `kr` instances on different paths — browser tabs have distinct titles

## Success Metrics

**Leading (immediate):**
- At a glance, you can tell which `kr` tab belongs to which project

**Lagging (2-4 weeks):**
- No longer having to check URLs or click around to identify the active project

**Failure signal:**
- The folder name is too generic (e.g., `docs`) to be useful — if this is a recurring complaint, consider a `--name` flag as a follow-up

## Implementation Hints

### Existing patterns to follow
- `Server` struct at `internal/server/server.go:17` — add `projectName string` field, set once in `New()` via `filepath.Base(rootPath)`
- `BacklogData`, `FolderData`, `DocumentData` at `internal/templates/templates.go:41-66` — add `ProjectName string` field to each
- `handleBacklog`, `handleFolder`, `handleDocument` at `internal/server/handlers.go` — set `ProjectName: s.projectName` when constructing each data struct
- `layout.html` at `internal/templates/layout.html:6,11` — replace hardcoded `kr` with `{{.ProjectName}}`

### Integration points
- **`internal/server/server.go`** — compute and store `projectName` once at server creation
- **`internal/templates/templates.go`** — add `ProjectName` field to all three page data structs
- **`internal/server/handlers.go`** — populate `ProjectName` in all three handlers
- **`internal/templates/layout.html`** — two references: `<title>` and nav brand `<a>`

### Technical risks
- None. `filepath.Base` is reliable for all path formats. No external dependencies, no schema changes, no new packages.

## Research Summary

Light research conducted directly in the codebase. Key findings:
- `Server.rootPath` is already available in all handlers via `s.rootPath` — no new state needed at the CLI level
- `layout.html` is the single template that controls both `<title>` and the nav brand — one file to update
- All three page data structs are plain Go structs with no embedding or interface constraints — adding a field is trivial
- `filepath.Base` correctly handles the path derivation for all cases (relative, absolute, trailing slash)

## Stories

1. **S-010: Show project folder name in nav bar and browser tab title**
   Derive `filepath.Base(rootPath)` once in `Server.New()`, add `ProjectName string` to `BacklogData`, `FolderData`, and `DocumentData`, populate it in all three handlers, and update `layout.html` to use `{{.ProjectName}}` in `<title>` and the nav brand link.
   - Acceptance: All page types (backlog, folder, document) show the served folder name in the nav bar brand and browser tab title. The brand link still navigates to `/`.
   - No dependencies.

## References

- Existing features: `docs/features/2026-03-07-live-docs-dashboard.md` (FEAT-001)
- Nav template: `internal/templates/layout.html:6,11`
- Template data structs: `internal/templates/templates.go:41-66`
- Handlers: `internal/server/handlers.go`
- Server struct: `internal/server/server.go:17`

## Origin

Feature spec created on 2026-03-13 through structured intake.
Original description: "I'd like to include in the navigation bar the name of the parent folder so I can visualize the project I'm watching. The reason is that I use to open the application for several projects at the same time and I use to get lost."
