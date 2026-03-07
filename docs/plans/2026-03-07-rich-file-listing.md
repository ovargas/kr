---
date: 2026-03-07
feature: FEAT-002
spec: docs/features/2026-03-07-rich-file-listing.md
status: approved
---

# Implementation Plan: Rich File Listing

## Overview

We're adding title and excerpt extraction to the folder file listing so users see document titles and previews instead of raw filenames. The approach is a lightweight line scanner (no goldmark dependency), wired through the existing `listFiles()` → `FileEntry` → `folder.html` pipeline. Ordered: extraction logic first, then data model + wiring, then template + CSS.

## Reference Implementation

The closest existing pattern is the folder listing pipeline:
- `internal/server/folders.go:51` — `listFiles()` builds `[]FileEntry` from directory entries
- `internal/templates/templates.go:48` — `FileEntry` struct defines the data shape
- `internal/templates/folder.html:1` — template renders the file list
- `internal/static/style.css:118` — `.file-list` CSS styles

This plan extends the same pipeline with richer data.

## Pre-conditions

Before starting implementation:
- [ ] Feature spec is approved: `docs/features/2026-03-07-rich-file-listing.md`
- [ ] All existing tests pass: `go test ./...`

---

## Phase 1: Title & Excerpt Extraction Logic

### Overview
Create a standalone extraction function with unit tests. This is the core new logic — a line scanner that reads raw markdown bytes, skips front matter, finds the H1 title, and extracts a plain-text excerpt from the following paragraph. No goldmark dependency.

### Step 1.1: Create extraction function
**File:** `internal/server/extract.go` (create)
**Pattern:** Standalone utility in the `server` package — follows the pattern of `folders.go` as a supporting file in the same package.

**What to do:**
Create a function `extractMeta(data []byte) (title string, excerpt string)` that:

1. Scans lines from `data`
2. Skips front matter: if the first line is `---`, skip lines until the closing `---`
3. Finds the first line starting with `# ` (after trimming) — everything after `# ` is the title
4. Collects subsequent non-empty lines that don't start with `#` into an excerpt paragraph
5. Strips basic inline markdown from the excerpt:
   - `**text**` and `__text__` → `text`
   - `*text*` and `_text_` → `text` (careful not to strip underscores in words)
   - `` `code` `` → `code`
   - `[text](url)` → `text`
   - `> ` blockquote prefix → remove
6. Truncates the excerpt to 200 characters at a word boundary, appending `...` if truncated
7. Returns `("", "")` if no H1 is found

**Details:**
- Use `bytes.Split` or `bufio.Scanner` to iterate lines — no need to read the entire file
- For inline markdown stripping, a few simple `strings.Replace` and one regex for `[text](url)` is sufficient
- Don't over-engineer the stripping — perfect is the enemy of good. Occasional markdown artifacts in edge cases are acceptable.

### Step 1.2: Unit tests for extraction
**File:** `internal/server/extract_test.go` (create)
**Pattern:** Follow `internal/renderer/renderer_test.go` for test structure — table-driven tests with string checks.

**What to do:**
Test cases to cover:

1. **Standard file:** front matter + H1 + paragraph → returns title and clean excerpt
2. **No front matter:** H1 + paragraph → returns title and excerpt
3. **No H1:** file with only content, no `# ` heading → returns `("", "")`
4. **Empty file:** returns `("", "")`
5. **H1 with no following content:** returns title and empty excerpt
6. **Excerpt with markdown syntax:** `**bold**`, `[link](url)`, `` `code` `` → stripped in excerpt
7. **Long excerpt:** paragraph > 200 chars → truncated at word boundary with `...`
8. **Front matter only, no content after:** returns `("", "")`
9. **Value line (blockquote after H1):** `> **Value:** Some text` → stripped to `Value: Some text`

### Phase 1 Verification

**Automated:**
- [ ] `go build ./internal/server/` — compiles without errors
- [ ] `go test ./internal/server/ -run TestExtract` — all extraction tests pass
- [ ] `go vet ./internal/server/` — no issues

**Stop here.** Verify Phase 1 before proceeding.

---

## Phase 2: Data Model & Wiring

### Overview
Extend `FileEntry` with `Title` and `Excerpt` fields, then update `listFiles()` to call `extractMeta()` for each file.

### Step 2.1: Extend FileEntry struct
**File:** `internal/templates/templates.go` (modify)
**Pattern:** Existing `FileEntry` at line 48.

**What to do:**
Add two fields to `FileEntry`:
- `Title string` — the H1 heading extracted from the file
- `Excerpt string` — the plain-text excerpt from the first section

### Step 2.2: Update listFiles to extract metadata
**File:** `internal/server/folders.go` (modify)
**Pattern:** Existing `listFiles()` at line 51.

**What to do:**
Inside the loop at line 58 where files are collected, after confirming the entry is a `.md` file:

1. Build the full file path: `filepath.Join(folderPath, e.Name())`
2. Read the file bytes: `os.ReadFile(path)`
3. Call `extractMeta(data)` to get title and excerpt
4. Populate `FileEntry{Name: e.Name(), Title: title, Excerpt: excerpt}`

If `os.ReadFile` fails for a file, skip the metadata (leave Title and Excerpt empty) — don't fail the whole listing.

### Step 2.3: Update handler test data and assertions
**File:** `internal/server/handlers_test.go` (modify)
**Pattern:** Existing `TestFolderListing` at line 56.

**What to do:**
The test at line 22 already creates a file with `# Hello` as the H1 and `World.` as content. Update `TestFolderListing` to also check that the response contains `"Hello"` (the title) and `"World."` (the excerpt). The existing check for `"test-doc.md"` should remain — the filename is still shown.

### Phase 2 Verification

**Automated:**
- [ ] `go build ./...` — compiles without errors
- [ ] `go test ./internal/server/` — all server tests pass (including updated folder listing test)
- [ ] `go test ./...` — full test suite passes

**Stop here.** Verify Phase 2 before proceeding.

---

## Phase 3: Template & CSS

### Overview
Update the folder listing template to show title, excerpt, and filename. Add CSS styles for the new layout.

### Step 3.1: Update folder template
**File:** `internal/templates/folder.html` (modify)
**Pattern:** Current template at line 1.

**What to do:**
Replace the simple `<li><a>{{.Name}}</a></li>` with a richer structure:

```
{{range .Files}}
<li>
    <a href="/{{$.FolderName}}/{{.Name}}">
        {{if .Title}}
            <span class="file-title">{{.Title}}</span>
        {{else}}
            <span class="file-title">{{.Name}}</span>
        {{end}}
    </a>
    {{if .Excerpt}}
        <p class="file-excerpt">{{.Excerpt}}</p>
    {{end}}
    {{if .Title}}
        <span class="file-name">{{.Name}}</span>
    {{end}}
</li>
{{end}}
```

Key behavior:
- If `Title` is present: show title as the link text, excerpt below, filename as secondary text
- If `Title` is empty: show filename as the link text (same as current behavior), no excerpt, no secondary filename

### Step 3.2: Add CSS styles for rich file listing
**File:** `internal/static/style.css` (modify)
**Pattern:** Existing `.file-list` styles at line 118.

**What to do:**
Add styles after the existing `.file-list` rules (after line 136):

- `.file-title` — font-weight 600, slightly larger than current link text
- `.file-excerpt` — smaller font size (~0.85rem), color #666, margin-top 0.25rem, line-height 1.4, no extra bottom margin
- `.file-name` — small font size (~0.78rem), color #999, display block, margin-top 0.15rem

Keep the existing `.file-list li` padding and border-bottom. The `<a>` wraps only the title, keeping the link click target clear.

### Phase 3 Verification

**Automated:**
- [ ] `go build ./...` — compiles without errors
- [ ] `go test ./...` — full test suite passes

**Manual:**
- [ ] Start `kr` with a docs directory containing date-prefixed `.md` files with H1 titles
- [ ] Navigate to a folder — verify titles and excerpts appear
- [ ] Verify a file without an H1 falls back to showing the filename
- [ ] Click a file entry — verify navigation to the document still works
- [ ] Check that excerpts are clean text (no `**`, `[]()`, or backticks showing)

**Stop here.** Verify Phase 3 before proceeding.

---

## Final Verification

**All automated checks:**
- [ ] Full test suite passes: `go test ./...`
- [ ] Vet passes: `go vet ./...`
- [ ] Build succeeds: `go build -o kr .`

**Manual testing:**
- [ ] Start `kr --path ./docs` and navigate to the `features/` folder
- [ ] Verify `2026-03-07-live-docs-dashboard.md` shows as "Live Docs Dashboard" with an excerpt
- [ ] Verify `2026-03-07-rich-file-listing.md` shows as "Rich File Listing" with an excerpt
- [ ] Create a `.md` file with no H1 heading — verify it shows the filename instead
- [ ] Verify existing Kanban board and document rendering are unaffected

**Definition of done alignment:**
- [ ] Folder file listing shows the document's H1 heading as the primary clickable label — Phase 3, Step 3.1
- [ ] The raw filename is displayed as secondary text below the title — Phase 3, Step 3.1
- [ ] A plain-text excerpt (~150-200 characters) from the first content section appears between title and filename — Phase 1 (extraction), Phase 3 (display)
- [ ] Files without an H1 heading fall back to showing the filename as the title, with no excerpt — Phase 1 (extraction returns empty), Phase 3 (template conditional)
- [ ] Existing folder navigation and document linking continue to work correctly — Phase 2 (handler test), Phase 3 (manual verification)

## Files Changed Summary

| File | Action | Phase | Notes |
|---|---|---|---|
| `internal/server/extract.go` | create | 1 | Title & excerpt extraction from raw markdown bytes |
| `internal/server/extract_test.go` | create | 1 | Unit tests for extraction function |
| `internal/templates/templates.go` | modify | 2 | Add Title, Excerpt fields to FileEntry |
| `internal/server/folders.go` | modify | 2 | Call extractMeta in listFiles loop |
| `internal/server/handlers_test.go` | modify | 2 | Assert title and excerpt in folder listing response |
| `internal/templates/folder.html` | modify | 3 | Rich listing with title, excerpt, filename |
| `internal/static/style.css` | modify | 3 | Styles for file-title, file-excerpt, file-name |

## Risks and Fallbacks

- **Performance on large folders:** Reading every `.md` file during listing adds I/O. For typical docs folders (5-20 files), this is negligible. Fallback: if someone reports a folder with 100+ files, add a limit (extract for top 50, rest show filename only).
- **Markdown stripping imperfections:** Edge cases in inline markdown may leave artifacts. Fallback: acceptable for V1 — iterate if specific patterns are reported. The excerpt is a preview, not a rendered document.

## References

- Feature spec: `docs/features/2026-03-07-rich-file-listing.md`
- File listing code: `internal/server/folders.go:51`
- File entry struct: `internal/templates/templates.go:48`
- Folder template: `internal/templates/folder.html`
- CSS styles: `internal/static/style.css:118`
- Handler tests: `internal/server/handlers_test.go:56`
