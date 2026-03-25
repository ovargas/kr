---
id: FEAT-004
date: 2026-03-25
status: draft
type: feature
research_level: light
yagni_verdict: build
plan: docs/plans/2026-03-25-content-search.md
tags: [search, navigation, navbar, regex]
---

# Content Search

> **Value:** Makes the entire docs directory instantly searchable from any page, eliminating the need to remember filenames or browse folders when you know what you're looking for.

## Problem

As documentation grows (features, plans, bugs, decisions, research), finding which file mentions a specific topic requires browsing through every folder or falling back to terminal `grep`. There is no way to search across all documents from within the viewer.

**Trigger:** The core viewer, rich file listing, and project naming are all in place. Search was explicitly deferred in FEAT-001 ("Search functionality — deferred, can be added as a separate feature later"). The docs directory is now large enough that browsing alone is insufficient.
**Current workaround:** `grep` in the terminal or click through folders hoping to find the right file. Both break flow and defeat the purpose of having a browser-based viewer.

## YAGNI Assessment

**Verdict:** BUILD IT

Search is the missing navigation primitive after folder browsing. The three toggle options (case sensitivity, whole word, regex) are cheap because they all collapse into the same `regexp` code path on the backend — the implementation cost of supporting all three is nearly identical to supporting just one. The 50-result cap avoids performance and UI complexity without needing pagination.

## Solution

### What we're building

1. **Search input in navbar:** A text input on the right side of the navigation bar, visible on all pages (backlog, folder, document). Three small toggle buttons next to it: case sensitive (Aa), whole word (W), and regex (.*). Submitting the form navigates to the search results page.

2. **Search endpoint:** `GET /search?q=<query>&case=<0|1>&word=<0|1>&regex=<0|1>` walks all `.md` files under `rootPath` recursively, checks each file for a match, and returns the list of matching files. Stops reading a file at the first match. Caps results at 50.

3. **Search results page:** Displays matching files in the same rich file-list format used by folder browsing (title + excerpt + filename). Each result includes the folder name so the user knows where the file lives. Results link to the correct document page (`/{folder}/{file}`). If more than 50 files match, a message indicates there are more results and the user should refine their search.

### How it works

1. User types a query in the navbar search input on any page
2. Optionally toggles case sensitivity, whole word, or regex mode
3. Presses Enter or clicks the search button
4. Browser navigates to `/search?q=term&case=0&word=0&regex=0`
5. Server receives the request, builds a `regexp.Regexp` from the query and options:
   - Plain text: `regexp.QuoteMeta(query)`, optionally wrapped with `\b...\b` for whole word
   - Regex mode: query used directly
   - Case insensitive: prepend `(?i)`
6. Server walks `rootPath` with `filepath.WalkDir`, filtering `.md` files
7. For each file, reads line-by-line with `bufio.Scanner`; stops at first match
8. For matching files, calls `extractMeta()` to get title and excerpt
9. Caps at 50 results; sets an overflow flag if limit is exceeded
10. Renders the search results template with the file list

### Visual concept

```
┌──────────────────────────────────────────────────────────────────────┐
│  docs  │ Backlog │ Bugs │ Features │ Plans │    [Aa][W][.*][ Search ]│
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Search results for "authentication"                                 │
│  12 files found                                                      │
│  ─────────                                                           │
│                                                                      │
│  features/                                                           │
│  Task Assignment Notifications                                       │
│  Guides new users through account setup and authentication           │
│  with contextual help...                                             │
│  2026-03-05-user-onboarding.md                                       │
│                                                                      │
│  decisions/                                                          │
│  Auth Middleware Rewrite                                              │
│  Session token storage compliance requirements for the new           │
│  authentication middleware...                                        │
│  2026-03-10-auth-middleware.md                                        │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

```
(When results exceed 50:)
┌──────────────────────────────────────────────────────────────────────┐
│  ...                                                                 │
│                                                                      │
│  ⚠ More than 50 results found. Refine your search to see more       │
│  specific results.                                                   │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

## Boundaries

### Explicitly NOT building
- **Match highlighting or context snippets** — results show the file's existing title+excerpt, not the matched line
- **Search indexing or caching** — every search scans the filesystem live
- **Pagination** — single capped result set at 50
- **Search-as-you-type / live suggestions** — submit-based only
- **Folder-scoped search** — always global across `rootPath`
- **Non-markdown file search** — `.md` files only
- **Relevance ranking** — files returned in filesystem walk order

### Rabbit holes to avoid
- **Debouncing or autocomplete** — this is a form submit, not an incremental search. Don't add JS complexity for real-time filtering.
- **Relevance scoring** — don't try to rank by match count, position, or file recency. Filesystem order is fine.
- **Content snippet at match site** — the existing `extractMeta()` excerpt is sufficient. Don't build a second extraction path for the matching line.
- **Concurrent file scanning** — sequential `filepath.WalkDir` is fast enough for typical doc directories (tens to low hundreds of files). Don't add goroutine pools.

## Definition of Done

**The feature is complete when:**

1. Every page (backlog, folder, document) shows a search input on the right side of the navbar with three toggle buttons (case sensitive, whole word, regex)
2. Submitting a query navigates to `/search?q=...` and displays matching `.md` files in the rich file-list format (title, excerpt, filename with folder prefix)
3. Each result links to the correct document page (`/{folder}/{file}`)
4. Results are capped at 50; if exceeded, a message says "More than 50 results. Refine your search."
5. Empty query or no matches shows an appropriate empty state
6. Toggle states (case, whole word, regex) are preserved in the URL and reflected in the UI on the results page
7. Invalid regex shows a user-friendly error message instead of crashing
8. Search scans recursively through all subdirectories under `rootPath`, only `.md` files

**Verification:**

Automated:
- [ ] `go test ./...` — all tests pass including new search logic tests
- [ ] `go build -o kr .` — builds without errors

Manual:
- [ ] Search for a known term — results appear with correct title+excerpt+filename
- [ ] Toggle case sensitivity — verify different results for "Bug" vs "bug"
- [ ] Toggle whole word — "plan" doesn't match "plans"
- [ ] Toggle regex — `feat.*` matches files containing "feature", "feat-001", etc.
- [ ] Enter invalid regex — error message shown, no crash
- [ ] Search yielding >50 results — overflow message displayed
- [ ] Search with no matches — empty state message
- [ ] Empty search — appropriate handling (empty state or redirect back)

## Success Metrics

**Leading (immediate):**
- Search returns correct results for known terms across all folders
- Search completes in under 1 second for a typical docs directory (~100 files)

**Lagging (2-4 weeks):**
- Search is the primary way to find documents when you know a keyword but not which folder it's in

**Failure signal:**
- Search is too slow to be useful (>3 seconds), or results are inaccurate enough that you still resort to terminal `grep`

## Implementation Hints

### Existing patterns to follow
- Handler pattern at `internal/server/handlers.go` — every handler: get nav items → build data struct → render template
- Template data structs at `internal/templates/templates.go:41-69` — each page type has its own struct with `ProjectName` and `NavItems`
- File listing reuse: `FileEntry{Name, Title, Excerpt}` at `internal/templates/templates.go:48-53`
- Title/excerpt extraction: `extractMeta()` at `internal/server/extract.go:13`
- Filesystem scanning: `listFiles()` at `internal/server/folders.go:51` and `filepath.WalkDir`
- Navbar layout: `internal/templates/layout.html:10-17` — flex layout, search form goes with `margin-left: auto`
- Route registration: `internal/server/server.go:60-63` — add `GET /search` before the catch-all

### Integration points
- **`internal/server/server.go`** — register `GET /search` route
- **`internal/server/handlers.go`** — new `handleSearch()` handler
- **`internal/server/search.go`** (new) — search logic: compile regex, walk files, match, collect results
- **`internal/templates/templates.go`** — new `SearchData` struct, new `SearchResultEntry` struct (extends `FileEntry` with `Folder`), new `RenderSearch()` method
- **`internal/templates/search.html`** (new) — search results template reusing file-list CSS
- **`internal/templates/layout.html`** — add search form to navbar
- **`internal/static/style.css`** — styles for search input, toggle buttons, overflow message
- **`internal/static/main.js`** — minimal JS for toggle button state management

### Search regex construction
- Plain text, case-insensitive: `regexp.Compile("(?i)" + regexp.QuoteMeta(query))`
- Plain text, case-sensitive: `regexp.Compile(regexp.QuoteMeta(query))`
- Whole word: wrap with `\b...\b`
- Regex mode: pass query directly to `regexp.Compile`
- Invalid regex: return error message to user, don't panic

### File scanning approach
- `filepath.WalkDir` under `rootPath`, skip hidden dirs (`.` prefix), filter `.md` files
- For each file: `bufio.Scanner` line-by-line, `re.MatchString(line)` — stop at first match
- Call `extractMeta()` on matching files to get title and excerpt
- Track result count; stop walking when limit (50) is exceeded

### Technical risks
- **Performance on large directories** — mitigated by first-match-per-file stopping and 50-result cap. Sequential scan is sufficient for typical docs directories.

## Research Summary

Light research conducted directly in the codebase. Key findings:
- All existing patterns (handler, template data, file listing, navbar) directly support adding a new page type
- `extractMeta()` is already extracted and reusable for search results
- `FileEntry` struct can be extended with a `Folder` field for search results (or a new struct wrapping it)
- Go stdlib `regexp`, `filepath.WalkDir`, and `bufio.Scanner` cover all search needs — no external libraries required
- The navbar flex layout naturally supports adding a right-aligned search form with `margin-left: auto`

## Stories

1. **S-011: Implement search logic — regex compilation and file scanning**
   Core engine. Takes a query string + options (case, whole word, regex), compiles a `regexp.Regexp`, walks `rootPath` recursively for `.md` files, scans each line-by-line stopping at first match, calls `extractMeta()` on matches, caps at 50 results. Returns a result list + overflow flag. Unit tested.
   - Acceptance: Given a directory with `.md` files, returns correct matches for plain text, case-sensitive/insensitive, whole word, and regex queries. Stops at 50 results and sets overflow flag. Invalid regex returns an error, doesn't panic.

2. **S-012: Add search handler and route**
   Integration. New `SearchData` and `SearchResultEntry` structs in templates. New `handleSearch()` handler that parses query params, calls search logic, builds template data, renders results. Register `GET /search` route.
   - Acceptance: `GET /search?q=term` returns an HTML page with matching files in rich file-list format. Toggle params (`case`, `word`, `regex`) are parsed and applied. Empty query shows empty state. Invalid regex shows error message.
   - Depends on: S-011

3. **S-013: Add search results template**
   New `search.html` template displaying results in the file-list format (title + excerpt + filename), grouped by folder. Shows result count, overflow message when >50, and empty state for no matches. New `RenderSearch()` method on `Templates`.
   - Acceptance: Search results page renders correctly with folder labels, rich file entries, result count, and overflow/empty states.
   - Depends on: S-012

4. **S-014: Add search input and toggles to navbar**
   UI. Add search form to `layout.html` (right side of navbar): text input, three toggle buttons (Aa, W, .*), submit. Minimal JS for toggle state (active/inactive class). CSS for search form, input, and toggle buttons matching navbar style. Form submits to `/search` with query params. On the search results page, the input is pre-filled and toggles reflect the current state.
   - Acceptance: Search input visible on all page types. Toggle buttons visually indicate active/inactive state. Submitting navigates to `/search?q=...&case=...&word=...&regex=...`. On the results page, input and toggles reflect the submitted query and options.
   - Depends on: S-013

## References

- Deferred from: `docs/features/2026-03-07-live-docs-dashboard.md` (FEAT-001, line 118: "Search functionality — deferred")
- File listing pattern: `internal/server/folders.go:51` (`listFiles`)
- Title/excerpt extraction: `internal/server/extract.go:13` (`extractMeta`)
- Template data structs: `internal/templates/templates.go:41-69`
- Navbar template: `internal/templates/layout.html:10-17`
- Handler pattern: `internal/server/handlers.go`
- Route registration: `internal/server/server.go:60-63`

## Origin

Feature spec created on 2026-03-25 through structured intake.
Original description: "I want to add a search input in the top bar so that when I enter a text, it can display a list of file where such text is found. The List should be shown in the same format the list of documents is shown when I browse a folder. The search must be available in all pages, it should be placed on the right side of the bar. No infrastructure service will be used to implement the search, the BE will search in the files within the provided path and return the list of files where the text is found. The search should support case sensitive/insensitive toggle, option to match whole words only, and option to use regular expressions. Stop reading a file as soon as we find the first match. Support max number of results (50). No pagination."
