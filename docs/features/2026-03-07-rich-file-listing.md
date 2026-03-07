---
id: FEAT-002
date: 2026-03-07
status: draft
type: feature
research_level: light
yagni_verdict: build
plan: docs/plans/2026-03-07-rich-file-listing.md
tags: [ui, navigation, file-listing, markdown]
---

# Rich File Listing

> **Value:** Makes folder browsing actually useful by showing document titles and excerpts instead of opaque date-prefixed filenames — you can tell what each document is about without clicking into it.

## Problem

The folder file listing shows raw filenames like `2026-03-07-live-docs-dashboard.md`. For date-prefixed files (plans, features, research), the filename doesn't convey what the document is about. You have to click into each file to discover its content, which makes browsing slow and frustrating.

**Trigger:** Using `kr` to browse docs folders — filenames are not descriptive enough to scan.
**Current workaround:** Click into each file to read the title, then go back. Inefficient when browsing folders with many files.

## YAGNI Assessment

**Verdict:** BUILD IT

The file listing is the primary navigation surface after the Kanban board. Making it scannable is a direct usability win with zero downside. The two pieces (title + excerpt) are both necessary: title tells you *what*, excerpt tells you *about what*. No trimming needed.

## Solution

### What we're building

1. **Title extraction:** Read the H1 heading (`# Title`) from each `.md` file and display it as the primary label in the file listing. The filename becomes secondary text.

2. **Excerpt extraction:** Read the first paragraph of content after the H1 heading (before the next heading or end of content). Display it as plain text, truncated to ~200 characters, below the title.

3. **Graceful fallback:** Files without an H1 heading show the filename as the title and no excerpt — identical to today's behavior.

### How it works

1. User clicks a folder in the navigation menu
2. Server scans the folder for `.md` files (existing behavior)
3. For each file, server reads the raw content and extracts:
   - The first `# Heading` line (after skipping front matter delimiters `---`)
   - The first paragraph of text after that heading (non-empty lines before the next `##` heading or blank line gap)
4. The file listing renders each entry with: title (from H1), excerpt (from first paragraph), and filename (as secondary text)
5. Clicking the entry navigates to the full rendered document (existing behavior)

### Visual concept

```
┌──────────────────────────────────────────────────────────────────┐
│  kr   │ Backlog │ Bugs │ Features │ Decisions │ Plans │ ...      │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Features                                                        │
│  ─────────                                                       │
│                                                                  │
│  Live Docs Dashboard                                             │
│  Gives you a real-time, browser-based control panel for          │
│  monitoring AI-agent documentation workflows...                  │
│  2026-03-07-live-docs-dashboard.md                               │
│                                                                  │
│  User Onboarding Flow                                            │
│  Guides new users through account setup and first project        │
│  creation with contextual help...                                │
│  2026-03-05-user-onboarding.md                                   │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Boundaries

### Explicitly NOT building
- Full markdown rendering for excerpts — plain text only, no styled HTML
- Front matter-based titles (e.g., a `title:` YAML field) — we use the H1 heading from the document content
- Caching or lazy loading — not needed at the scale of typical docs folders (5-20 files)
- Search or filtering of documents — separate feature if needed later

### Rabbit holes to avoid
- **Using goldmark for extraction** — tempting because it's already in the project, but a simple line scanner is much lighter. We don't need to parse the full AST just to grab a title and first paragraph.
- **Over-engineering fallback logic** — if there's no H1, show the filename. Don't try to infer titles from filenames by parsing dates and converting kebab-case.
- **Markdown stripping in excerpts** — simple inline syntax removal (bold, italic, links, backticks) is sufficient. Don't try to handle every markdown edge case.

## Definition of Done

**The feature is complete when:**

1. Folder file listing shows the document's H1 heading as the primary clickable label
2. The raw filename is displayed as secondary text below or beside the title
3. A plain-text excerpt (~150-200 characters) from the first content section appears between the title and filename
4. Files without an H1 heading fall back to showing the filename as the title, with no excerpt
5. Existing folder navigation and document linking continue to work correctly

**Verification:**

Automated:
- [ ] `go test ./...` — all tests pass, including new tests for title/excerpt extraction
- [ ] `go build -o kr .` — binary builds without errors

Manual:
- [ ] Start `kr` with a docs directory containing date-prefixed `.md` files
- [ ] Navigate to a folder — verify titles and excerpts are shown instead of raw filenames
- [ ] Verify a file without an H1 heading falls back to showing the filename
- [ ] Click a file entry — verify it navigates to the full rendered document
- [ ] Verify excerpts are clean plain text (no markdown syntax like `**`, `[]()`  leaking through)

## Success Metrics

**Leading (immediate):**
- File listings are scannable — you can tell what each document is about without clicking into it

**Lagging (2-4 weeks):**
- Browsing folders feels natural — no need to remember what date-prefixed filenames map to

**Failure signal:**
- Excerpts are unhelpful (too short, garbled markdown syntax, or irrelevant boilerplate leaking through), causing you to still click into files to understand them

## Implementation Hints

### Existing patterns to follow
- `listFiles()` at `internal/server/folders.go:51` — currently builds `[]FileEntry` with only `Name`. Extend to read each file and extract title + excerpt.
- `FileEntry` at `internal/templates/templates.go:48` — add `Title` and `Excerpt` fields.
- `folder.html` at `internal/templates/folder.html` — update template to render title, excerpt, and filename.

### Integration points
- **`internal/server/folders.go`** — `listFiles()` needs to accept `folderPath` (already does) and read file contents for extraction
- **`internal/templates/templates.go`** — `FileEntry` struct needs new fields
- **`internal/templates/folder.html`** — template needs updated markup for rich listing

### Title/excerpt extraction approach
- Lightweight line scanner, no goldmark dependency:
  1. Read file bytes
  2. Skip front matter (content between `---` delimiters at the start)
  3. Find the first line starting with `# ` — that's the title
  4. Collect subsequent non-empty, non-heading lines as the excerpt paragraph
  5. Strip basic inline markdown syntax (`**`, `*`, `` ` ``, `[text](url)` → `text`)
  6. Truncate to ~200 characters at a word boundary

### Technical risks
- **Performance on large folders** — reading every `.md` file for listing. Mitigated: typical folders have 5-20 files, and we're only reading the first ~50 lines, not the whole file.

## Research Summary

Light research conducted directly in the codebase. Key findings:
- `listFiles()` and `FileEntry` are the only touchpoints for folder listing data
- The renderer (`goldmark`) is available but overkill for title extraction — a line scanner is simpler and faster
- The `folder.html` template is minimal (8 lines) and easy to extend
- No caching infrastructure exists or is needed at this scale

## Stories

1. **S-007: Add title and excerpt extraction for markdown files**
   Core logic. A new function that reads a `.md` file's raw bytes and returns the H1 title and a plain-text excerpt from the first content section. Includes skipping front matter, stripping basic markdown syntax, and truncating to ~200 characters at a word boundary. Unit tested.
   - Acceptance: Given a markdown file with front matter + H1 + content, returns the correct title and a clean ~200-char excerpt. Given a file with no H1, returns empty title and no excerpt.

2. **S-008: Extend file listing to include title and excerpt**
   Integration. Update `FileEntry` struct with `Title` and `Excerpt` fields. Update `listFiles()` to call the extraction function for each file. Wire it through `handleFolder()`.
   - Acceptance: `FileEntry` populated with title and excerpt. Files without H1 have empty title/excerpt. Existing folder handler returns enriched data.
   - Depends on: S-007

3. **S-009: Update folder template for rich file listing**
   UI. Update `folder.html` to render title as primary label, excerpt below it, filename as secondary text. Style with existing CSS conventions.
   - Acceptance: Folder page shows title + excerpt + filename for each file. Falls back to filename-only display when no title extracted. Links still navigate to the correct document.
   - Depends on: S-008

## References

- Existing feature: `docs/features/2026-03-07-live-docs-dashboard.md` (FEAT-001)
- File listing code: `internal/server/folders.go:51` (`listFiles`)
- File entry struct: `internal/templates/templates.go:48` (`FileEntry`)
- Folder template: `internal/templates/folder.html`

## Origin

Feature spec created on 2026-03-07 through structured intake.
Original description: "In order to improve a bit the experience, In the list of files, I'd like to read the title of the file instead of the filename. For example, instead of '2026-03-07-live-docs-dashboard.md', it would be nice to read 'Live Docs Dashboard'. Also, read an extract of the immediately following section, so I can have a better idea of what the file is about."
