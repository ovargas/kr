---
date: 2026-03-13
feature: FEAT-003
spec: docs/features/2026-03-13-project-name-in-navbar.md
status: approved
---

# Implementation Plan: Project Name in Navigation Bar

## Overview

Add the served folder's name to the nav bar brand and browser tab title across all page types. The project name is derived once from `filepath.Base(rootPath)` at server creation and threaded through the existing template data structs. Single phase — four files, no new dependencies.

## Reference Implementation

The closest pattern is how `NavItems` is currently threaded through every page:
- `internal/server/server.go:27` — `New()` initializes server state
- `internal/server/handlers.go:15-21` — each handler builds a data struct with `NavItems`
- `internal/templates/templates.go:42-66` — all three data structs carry `NavItems []NavItem`
- `internal/templates/layout.html:12-16` — layout template reads `{{range .NavItems}}`

This plan follows the exact same pattern: add a field to the data structs, populate it in each handler, consume it in the layout template.

## Pre-conditions

- [ ] Feature spec approved: `docs/features/2026-03-13-project-name-in-navbar.md`

---

## Phase 1: Thread Project Name from Server to Templates

### Overview

Store the project name on the `Server` struct, add `ProjectName` to all three template data structs, populate it in all handlers, and update the layout template. One phase because there are no intermediate verifiable states — the change is atomic.

### Step 1.1: Add `projectName` field to `Server`

**File:** `internal/server/server.go` (modify)
**Pattern:** Follows existing field pattern at line 18 (`rootPath string`)

**What to do:**
- Add `projectName string` field to the `Server` struct
- In `New()`, after setting `rootPath`, set `projectName: filepath.Base(rootPath)`
- No new imports needed — `path/filepath` is already imported

### Step 1.2: Add `ProjectName` to template data structs

**File:** `internal/templates/templates.go` (modify)
**Pattern:** Follows existing `NavItems` field on each struct

**What to do:**
- Add `ProjectName string` field to `BacklogData` (line 42-45)
- Add `ProjectName string` field to `FolderData` (line 55-59)
- Add `ProjectName string` field to `DocumentData` (line 62-66)

### Step 1.3: Populate `ProjectName` in all handlers

**File:** `internal/server/handlers.go` (modify)
**Pattern:** Follows how `NavItems` is set in each handler's data struct

**What to do:**
- In `handleBacklog` (line 21): add `ProjectName: s.projectName` to `BacklogData`
- In `handleFolder` (line 91): add `ProjectName: s.projectName` to `FolderData`
- In `handleDocument` (line 137): add `ProjectName: s.projectName` to `DocumentData`

### Step 1.4: Use `ProjectName` in the layout template

**File:** `internal/templates/layout.html` (modify)

**What to do:**
- Line 6: change `<title>kr</title>` to `<title>{{.ProjectName}}</title>`
- Line 11: change `<a href="/" class="nav-brand">kr</a>` to `<a href="/" class="nav-brand">{{.ProjectName}}</a>`

### Phase 1 Verification

**Automated:**
- [ ] `go build -o kr .` — binary builds without errors
- [ ] `go test ./...` — all existing tests pass
- [ ] `go vet ./...` — no issues

**Manual:**
- [ ] Run `kr --path ./docs` — nav brand shows "docs", browser tab title shows "docs"
- [ ] Navigate to a folder (e.g., `/features/`) — "docs" persists in nav and title
- [ ] Navigate to a document — "docs" persists in nav and title
- [ ] Run a second instance with a different path — browser tabs show distinct names

---

## Final Verification

**All automated checks:**
- [ ] `go build -o kr .`
- [ ] `go test ./...`
- [ ] `go vet ./...`

**Manual testing:**
- [ ] Start with `--path ./docs` — brand and title show "docs"
- [ ] Start with `--path /tmp/myproject` — brand and title show "myproject"
- [ ] Start with `--path .` from a named directory — shows the directory name
- [ ] Brand link (`/`) still navigates to the Kanban board

**Definition of done alignment:**
- [ ] DoD 1 (nav brand shows folder name on all pages) — Phase 1, Steps 1.1-1.4
- [ ] DoD 2 (browser tab title shows folder name) — Phase 1, Step 1.4
- [ ] DoD 3 (brand link navigates to `/`) — unchanged, verified manually

## Files Changed Summary

| File | Action | Phase | Notes |
|---|---|---|---|
| `internal/server/server.go` | modify | 1 | Add `projectName` field, set in `New()` |
| `internal/templates/templates.go` | modify | 1 | Add `ProjectName` to 3 data structs |
| `internal/server/handlers.go` | modify | 1 | Populate `ProjectName` in 3 handlers |
| `internal/templates/layout.html` | modify | 1 | Use `{{.ProjectName}}` in title and brand |

## Risks and Fallbacks

- No meaningful risks. All changes are additive (new struct field) or substitutional (template text). No behavior changes to existing logic.

## References

- Feature spec: `docs/features/2026-03-13-project-name-in-navbar.md`
- Server struct: `internal/server/server.go:17`
- Template data structs: `internal/templates/templates.go:42-66`
- Handlers: `internal/server/handlers.go:14-146`
- Layout template: `internal/templates/layout.html:6,11`
