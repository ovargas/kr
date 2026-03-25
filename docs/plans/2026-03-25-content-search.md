---
date: 2026-03-25
feature: FEAT-004
spec: docs/features/2026-03-25-content-search.md
status: approved
---

# Implementation Plan: Content Search

## Overview

We're adding a full-text content search to kr — a search input in the navbar, a backend that walks `.md` files matching a regex, and a results page reusing the rich file-list format. The plan follows the same layered pattern as the existing features: pure logic first (search engine), then template data structs, then the HTTP handler, and finally the navbar UI.

## Reference Implementation

The closest existing pattern is the folder listing feature:
- `internal/server/folders.go:51` — `listFiles()` scans a directory for `.md` files, reads each, calls `extractMeta()`
- `internal/server/handlers.go:64` — `handleFolder()` gets nav items, builds `FolderData`, renders template
- `internal/templates/templates.go:56` — `FolderData` struct with `ProjectName`, `NavItems`, `Files`
- `internal/templates/folder.html` — renders `FileEntry` items as a list with title, excerpt, filename
- `internal/server/handlers_test.go:13` — `setupTestServer()` creates temp dir with `.md` files for HTTP tests

This plan follows the same structure with adaptations for: recursive file walking, regex matching, search-specific query params, and a navbar search form.

## Pre-conditions

Before starting implementation:
- [x] Feature spec is approved: `docs/features/2026-03-25-content-search.md`
- No external setup or dependencies needed — pure Go stdlib

---

## Phase 1: Search Logic — Regex Compilation and File Scanning

### Overview
Build the core search engine as a standalone, testable unit. No HTTP, no templates — just a function that takes a query + options, walks the filesystem, and returns matching files. This is the riskiest piece (regex edge cases, filesystem walking, performance), so it goes first.

### Step 1.1: Create search options and result types
**File:** `internal/server/search.go` (create)
**Pattern:** Follow `internal/server/extract.go` for file structure within the `server` package

**What to do:**
Define the search types and the regex compilation function:

- `SearchOptions` struct: `Query string`, `CaseSensitive bool`, `WholeWord bool`, `UseRegex bool`
- `SearchResult` struct: `Folder string`, `Name string`, `Title string`, `Excerpt string` — extends the `FileEntry` concept with a `Folder` field
- `SearchOutcome` struct: `Results []SearchResult`, `Overflow bool` — wraps the result list with an overflow flag
- `compilePattern(opts SearchOptions) (*regexp.Regexp, error)` function:
  - If `UseRegex` is false: `regexp.QuoteMeta(opts.Query)` to escape special chars
  - If `WholeWord` is true: wrap with `\b...\b`
  - If `CaseSensitive` is false: prepend `(?i)`
  - If `UseRegex` is true: use `opts.Query` directly (no QuoteMeta)
  - Return `regexp.Compile(pattern)` — invalid regex returns an error

### Step 1.2: Implement file matching function
**File:** `internal/server/search.go` (continue)
**Pattern:** Follow `internal/server/folders.go:51` (`listFiles`) for filesystem reading and `extractMeta()` usage

**What to do:**
Create `matchesFile(filePath string, re *regexp.Regexp) bool`:
- Open the file with `os.Open`
- Wrap in `bufio.Scanner` for line-by-line reading
- For each line, call `re.MatchString(line)`
- Return `true` at first match (short-circuit)
- Return `false` if no lines match or on read error

### Step 1.3: Implement the recursive search walk
**File:** `internal/server/search.go` (continue)
**Pattern:** Follow `internal/server/folders.go:16` (`scanFolders`) for directory traversal, adapted to recursive walking

**What to do:**
Create `searchFiles(rootPath string, opts SearchOptions, maxResults int) (SearchOutcome, error)`:
- Call `compilePattern(opts)` — if error, return it (invalid regex)
- Use `filepath.WalkDir(rootPath, ...)` to walk recursively
- In the walk function:
  - Skip hidden directories (name starts with `.`)
  - Skip non-`.md` files
  - For each `.md` file, call `matchesFile(fullPath, re)`
  - On match: compute `folder` as the relative directory from `rootPath` (e.g., `filepath.Rel(rootPath, dir)`), read file bytes, call `extractMeta()` for title+excerpt, append a `SearchResult{Folder, Name, Title, Excerpt}`
  - Track result count; when count exceeds `maxResults`, set `Overflow = true` and return early (use `filepath.SkipAll` to stop walking)
- Return `SearchOutcome{Results, Overflow}`

**Edge cases:**
- Files directly in `rootPath` (no subfolder): `Folder` should be `""` or `.`
- Empty query: return immediately with empty results (don't walk)

### Step 1.4: Unit tests for search logic
**File:** `internal/server/search_test.go` (create)
**Pattern:** Follow `internal/server/extract_test.go` for test structure — table-driven tests with `t.TempDir()`

**What to do:**
Write tests covering:
- `compilePattern` tests:
  - Plain text, case-insensitive (default) — matches regardless of case
  - Plain text, case-sensitive — only matches exact case
  - Whole word — `"plan"` doesn't match `"plans"`
  - Regex mode — `"feat.*"` matches `"feature"`
  - Invalid regex — returns error
- `matchesFile` tests:
  - File with matching content — returns true
  - File without matching content — returns false
- `searchFiles` integration tests:
  - Create temp dir with multiple folders and `.md` files with known content
  - Search for a term present in some files — verify correct results with folder, name, title, excerpt
  - Search exceeding max results — verify overflow flag
  - Empty query — returns empty results
  - Case sensitivity toggle — different results for "Bug" vs "bug"
  - Whole word toggle — "plan" vs "plans"
  - Regex search — `"^## Problem"` matches files with that heading

### Phase 1 Verification

**Automated:**
- [ ] `go test ./internal/server/ -run TestCompilePattern` — regex compilation tests pass
- [ ] `go test ./internal/server/ -run TestMatchesFile` — file matching tests pass
- [ ] `go test ./internal/server/ -run TestSearchFiles` — integration tests pass
- [ ] `go vet ./internal/server/` — no issues

**Stop here.** Verify Phase 1 before proceeding.

---

## Phase 2: Template Data Structs and Search Results Template

### Overview
Add the data structures and HTML template needed to render search results. This wires the search engine output to the rendering layer without touching HTTP yet.

### Step 2.1: Add SearchData struct and RenderSearch method
**File:** `internal/templates/templates.go` (modify)
**Pattern:** Follow `FolderData` at line 56 and `RenderFolder` at line 115

**What to do:**
- Add `SearchResultEntry` struct: `Folder string`, `Name string`, `Title string`, `Excerpt string`
- Add `SearchData` struct: `ProjectName string`, `NavItems []NavItem`, `Query string`, `CaseSensitive bool`, `WholeWord bool`, `UseRegex bool`, `Results []SearchResultEntry`, `Overflow bool`, `Error string`, `TotalFound int`
- Add a `search` field to the `Templates` struct (type `*template.Template`)
- In `New()`, parse `search.html` the same way as `folder.html`: `template.Must(layout.Clone()).ParseFS(templateFS, "search.html")`
- Add `RenderSearch(w io.Writer, data SearchData) error` method following the pattern of `RenderFolder`

### Step 2.2: Create search results template
**File:** `internal/templates/search.html` (create)
**Pattern:** Follow `internal/templates/folder.html` for the file-list rendering structure

**What to do:**
Create the template with:
- Header: "Search results for `<query>`" with result count (e.g., "12 files found")
- If `Error` is non-empty: display error message (e.g., "Invalid regular expression: <error>") styled as `.search-error`
- If no results and no error: empty state "No files match your search"
- Results list using the same `.file-list` CSS class as folder listing:
  - Group by folder — show folder name as a subheading before its entries
  - Each entry: link to `/{Folder}/{Name}` with title (or filename fallback), excerpt, and filename
  - Follow the exact same HTML structure as `folder.html` for each file entry
- If `Overflow` is true: a message at the bottom: "More than 50 results found. Refine your search to see more specific results." styled as `.search-overflow`

**Template structure:**
```
{{define "content"}}
<h1>Search results for "{{.Query}}"</h1>
{{if .Error}}...error display...
{{else if .Results}}
  <p class="search-count">{{.TotalFound}} files found{{if .Overflow}} (showing first 50){{end}}</p>
  <ul class="file-list">
    {{range .Results}}...file entries with folder label...{{end}}
  </ul>
  {{if .Overflow}}<p class="search-overflow">...</p>{{end}}
{{else}}
  <p class="empty-state">No files match your search</p>
{{end}}
{{end}}
```

### Phase 2 Verification

**Automated:**
- [ ] `go build ./internal/templates/` — templates package compiles
- [ ] `go vet ./internal/templates/` — no issues
- [ ] `go test ./internal/server/` — existing tests still pass (template parsing doesn't break)

**Stop here.** Verify Phase 2 before proceeding.

---

## Phase 3: Search Handler and Route Registration

### Overview
Wire the search logic to the HTTP layer. Add the handler that parses query parameters, calls `searchFiles()`, and renders the search template.

### Step 3.1: Add handleSearch handler
**File:** `internal/server/handlers.go` (modify)
**Pattern:** Follow `handleFolder` at line 64 — same shape: get nav items, build data struct, render template

**What to do:**
Add `func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request)`:
- Parse query params: `q` (search query), `case` (0/1), `word` (0/1), `regex` (0/1)
- Build `SearchOptions` from params
- Get nav items via `scanFolders(s.rootPath)`
- Build `SearchData` with `ProjectName`, `NavItems`, `Query`, toggle states
- If query is empty: render with empty results (no search performed)
- Call `searchFiles(s.rootPath, opts, 50)`:
  - On error (invalid regex): set `SearchData.Error` to the error message, render
  - On success: map `SearchOutcome.Results` to `[]SearchResultEntry`, set `TotalFound` and `Overflow`
- Render via `s.tmpl.RenderSearch(w, data)`

### Step 3.2: Register the search route
**File:** `internal/server/server.go` (modify)
**Pattern:** Follow route registration at line 60-63

**What to do:**
Add the search route BEFORE the catch-all route:
```
mux.HandleFunc("GET /search", s.handleSearch)
```
Place it between the `/events` route (line 61) and the static file handler (line 62). The catch-all `GET /` at line 63 must remain last.

### Step 3.3: HTTP handler tests
**File:** `internal/server/handlers_test.go` (modify)
**Pattern:** Follow `TestFolderListing` at line 56 — same structure: setup server, make request, check response

**What to do:**
Add tests:
- `TestSearchBasic`: Create temp dir with `.md` files containing known text. `GET /search?q=known-term` → 200, response contains matching file title
- `TestSearchNoResults`: `GET /search?q=nonexistent-term` → 200, response contains "No files match"
- `TestSearchEmptyQuery`: `GET /search?q=` → 200, response contains empty state
- `TestSearchInvalidRegex`: `GET /search?q=[invalid&regex=1` → 200, response contains error message (not 500)
- `TestSearchCaseSensitive`: same term different case, verify toggle changes results
- `TestSearchOverflow`: create >50 `.md` files matching, verify overflow message

### Phase 3 Verification

**Automated:**
- [ ] `go test ./internal/server/ -run TestSearch` — all new search handler tests pass
- [ ] `go test ./internal/server/` — all existing tests still pass
- [ ] `go vet ./...` — no issues

**Manual:**
- [ ] Start `kr` with a docs directory, navigate to `/search?q=test` in browser — results page renders

**Stop here.** Verify Phase 3 before proceeding.

---

## Phase 4: Navbar Search Input and Toggle Buttons

### Overview
Add the search form to the shared navbar layout so it appears on every page. Add minimal JS for toggle button state and CSS for styling.

### Step 4.1: Add search form to navbar
**File:** `internal/templates/layout.html` (modify)
**Pattern:** Existing navbar structure at lines 10-17

**What to do:**
Add a search form after `.nav-links` div, inside the `.navbar`:
- A `<form>` with `action="/search"` method="GET", class `search-form`
- Style with `margin-left: auto` to push it to the right side of the navbar
- Three toggle buttons (type `button`, not submit): case sensitive (`Aa`), whole word (`W`), regex (`.*`)
  - Each button has a `data-param` attribute (`case`, `word`, `regex`) and toggles an `active` class
  - Each has a corresponding hidden `<input type="hidden" name="case" value="0">` that JS updates
- A text input (`name="q"`, placeholder "Search...")
- The form submits naturally via Enter key (default form behavior, no submit button needed visually — or a small search icon button)

**Pre-fill on search results page:**
- The `SearchData` struct already includes `Query`, `CaseSensitive`, `WholeWord`, `UseRegex`
- The layout template needs access to search state. Two options:
  - Option A: Add search fields (`Query`, `CaseSensitive`, `WholeWord`, `UseRegex`) to every data struct (BacklogData, FolderData, DocumentData, SearchData)
  - Option B: Use a common interface or embed a shared struct
  - **Chosen approach:** Add `SearchQuery string`, `SearchCase bool`, `SearchWord bool`, `SearchRegex bool` fields to all four data structs. Only `SearchData` populates them with actual values; the other three leave them as zero values. The layout template uses these to pre-fill the form. This follows the same pattern used for `ProjectName` — a field on every struct used by the shared layout.

### Step 4.2: Update all template data structs with search state fields
**File:** `internal/templates/templates.go` (modify)
**Pattern:** Follow how `ProjectName` was added to all three structs in FEAT-003

**What to do:**
Add four fields to `BacklogData`, `FolderData`, `DocumentData`, and `SearchData`:
- `SearchQuery string`
- `SearchCase bool`
- `SearchWord bool`
- `SearchRegex bool`

Only `handleSearch` needs to populate these. The existing handlers (`handleBacklog`, `handleFolder`, `handleDocument`) leave them as zero values, which means the form renders with an empty input and all toggles off — the correct default.

### Step 4.3: Add CSS for search form and toggles
**File:** `internal/static/style.css` (modify)
**Pattern:** Follow existing navbar styles at lines 12-50

**What to do:**
Add styles for:
- `.search-form`: `display: flex; align-items: center; gap: 0.25rem; margin-left: auto;`
- `.search-input`: text input styled to fit the navbar (dark background, light text, compact size, ~180px width)
- `.search-toggle`: small toggle buttons styled to match the navbar (transparent background, border, light text when inactive, highlighted background when active)
- `.search-toggle.active`: visually distinct active state (e.g., `background: rgba(255,255,255,0.2); color: #fff`)
- `.search-count`: result count text, similar to `.empty-state` but not italic
- `.search-overflow`: warning message style (subtle warning color)
- `.search-error`: error message style (subtle error color)
- `.search-folder-label`: folder name label above grouped results (styled like a subheading)
- Responsive: on narrow screens, search input can shrink or wrap

### Step 4.4: Add JS for toggle button state
**File:** `internal/static/main.js` (modify)
**Pattern:** Extend the existing IIFE at line 1

**What to do:**
Add within the existing IIFE:
- Select all `.search-toggle` buttons
- On click: toggle `active` class, update the corresponding hidden input value (0 ↔ 1)
- On page load: if hidden inputs have value `1` (pre-filled from URL), add `active` class to corresponding button
- Keep it minimal — no framework, plain DOM manipulation

### Phase 4 Verification

**Automated:**
- [ ] `go build -o kr .` — binary builds
- [ ] `go test ./...` — all tests pass
- [ ] `go vet ./...` — no issues

**Manual:**
- [ ] Start `kr`, verify search input appears in navbar on backlog page
- [ ] Navigate to a folder — search input still visible
- [ ] Navigate to a document — search input still visible
- [ ] Click toggle buttons — they visually toggle active/inactive
- [ ] Type a query, press Enter — navigates to `/search?q=...` with correct params
- [ ] On results page — input is pre-filled, active toggles are highlighted
- [ ] Toggle case sensitivity, re-search — different results for "Bug" vs "bug"
- [ ] Toggle whole word, re-search — "plan" doesn't match "plans"
- [ ] Toggle regex, search `^## Problem` — matches files with that heading
- [ ] Enter invalid regex with regex toggle on — error message displayed
- [ ] Search with >50 matches — overflow message shown

**Stop here.** Verify Phase 4 before proceeding.

---

## Final Verification

**All automated checks:**
- [ ] Full test suite passes: `go test ./...`
- [ ] Linting passes: `golangci-lint run ./...` (if installed)
- [ ] Vet passes: `go vet ./...`
- [ ] Build succeeds: `go build -o kr .`

**Manual testing:**
- [ ] Start `kr --path ./docs`, search for "authentication" — verify results appear correctly
- [ ] Verify each result links to the correct document page
- [ ] Test all three toggle combinations
- [ ] Verify search works from every page type (backlog, folder, document)
- [ ] Verify existing features (folder navigation, document viewing, Kanban board, live reload) still work

**Definition of done alignment:**
- [ ] Search input on all pages (backlog, folder, document) with toggles — Phase 4, Step 4.1
- [ ] Submit navigates to `/search?q=...` with rich file-list results — Phase 3, Step 3.1
- [ ] Each result links to correct document page — Phase 2, Step 2.2
- [ ] Results capped at 50 with overflow message — Phase 1, Step 1.3 + Phase 2, Step 2.2
- [ ] Empty query / no matches shows appropriate empty state — Phase 3, Step 3.1 + Phase 2, Step 2.2
- [ ] Toggle states preserved in URL and reflected in UI — Phase 4, Steps 4.1-4.4
- [ ] Invalid regex shows user-friendly error — Phase 3, Step 3.1 + Phase 2, Step 2.2
- [ ] Recursive search of all `.md` files — Phase 1, Step 1.3

## Files Changed Summary

| File | Action | Phase | Notes |
|---|---|---|---|
| `internal/server/search.go` | create | 1 | Search logic: regex compilation, file matching, recursive walk |
| `internal/server/search_test.go` | create | 1 | Unit tests for search logic |
| `internal/templates/templates.go` | modify | 2, 4 | Add SearchData, SearchResultEntry, RenderSearch, search state fields on all structs |
| `internal/templates/search.html` | create | 2 | Search results page template |
| `internal/server/handlers.go` | modify | 3 | Add handleSearch handler |
| `internal/server/server.go` | modify | 3 | Register GET /search route |
| `internal/server/handlers_test.go` | modify | 3 | HTTP tests for search endpoint |
| `internal/templates/layout.html` | modify | 4 | Add search form to navbar |
| `internal/static/style.css` | modify | 4 | Search form, toggle, overflow styles |
| `internal/static/main.js` | modify | 4 | Toggle button state management |

## Risks and Fallbacks

- **Performance on large directories:** Sequential `filepath.WalkDir` + line-by-line scanning could be slow for hundreds of files. Mitigated by first-match short-circuit and 50-result cap. Fallback: if profiling shows issues, add a goroutine pool (but don't prematurely optimize).
- **Regex denial of service:** A malicious or accidental regex (e.g., `(a+)+`) could cause catastrophic backtracking. Mitigated by the line-by-line scanning approach (each line is bounded in length). Fallback: add a timeout context if needed.
- **Toggle state UX:** Hidden inputs + JS toggles could break if JS is disabled. Acceptable for a local dev tool. Fallback: the form still submits with default values (all toggles off).

## References

- Feature spec: `docs/features/2026-03-25-content-search.md`
- Folder listing pattern: `internal/server/folders.go:51`, `internal/server/handlers.go:64`
- Template data pattern: `internal/templates/templates.go:56`
- File list template: `internal/templates/folder.html`
- Test pattern: `internal/server/handlers_test.go:13`, `internal/server/extract_test.go`
- Navbar: `internal/templates/layout.html:10-17`
- CSS: `internal/static/style.css:12-50`
