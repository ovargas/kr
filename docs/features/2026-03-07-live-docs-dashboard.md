---
id: FEAT-001
date: 2026-03-07
status: draft
type: feature
epic: EPIC-001
hub_decisions: []
research_level: light
yagni_verdict: build
plan: docs/plans/2026-03-07-live-docs-dashboard.md
tags: [cli, web-server, kanban, markdown, live-reload]
---

# Live Documentation Dashboard

> **Value:** Gives you a real-time, browser-based control panel for monitoring AI-agent documentation workflows — replacing manual terminal file reading with a live Kanban board and navigable doc viewer.

## Problem

AI agents continuously update project documentation (backlog, features, plans, bugs, decisions), but there's no way to monitor this activity in real time. You read raw markdown in terminals, manually refresh, and mentally assemble project state from scattered files. The backlog — a structured file that maps naturally to a Kanban board — is consumed as raw text with no visual overview.

**Trigger:** Running multiple AI agents that update docs continuously. Flying blind without a live viewer is blocking productivity.
**Current workaround:** `cat` / editor to read files, manually navigate between referenced documents, refresh to see changes. Insufficient because it breaks flow and misses real-time updates.

## YAGNI Assessment

**Verdict:** BUILD IT

Every proposed capability maps to the stated problem. The scope is already the minimum viable version: Kanban board (core value), folder navigation (necessary for browsing), markdown rendering (the point of a docs viewer), and live reload (the differentiator). Nothing to trim.

## Solution

### What we're building

1. **CLI entry point:** `kr --port <port> --path <dir>` starts a local web server. Port defaults to a random available port; path defaults to current directory. Prints `http://localhost:<port>` to stdout on startup.

2. **Kanban backlog view (root URL):** Parses `backlog.md` into 4 columns — Inbox, Ready, Doing, Done — displayed left to right in that order. Each backlog item is rendered as a card showing parsed fields (id, title, feature, service, depends, plan, spec, epic, etc.). Any field value matching `docs/{folder}/{file}.md` becomes a clickable link to the rendered page within the viewer.

3. **Folder navigation menu:** A persistent top menu bar listing all directories under `--path`. Known folders (bugs, features, decisions, plans, research, reviews, handoffs) appear first in a fixed order. Unknown folders appear after, sorted alphabetically. Clicking the app name/home returns to the Kanban board.

4. **Folder file listing:** Clicking a folder in the menu shows a list of `.md` files in that folder, sorted by filename (newest first for date-prefixed files).

5. **Markdown document viewer:** Clicking a file renders it as styled HTML via goldmark. Front matter (YAML) is extracted and displayed as a metadata section above the content. Standard markdown elements are properly formatted: headings, lists, links, code blocks, tables, emphasis.

6. **Live auto-refresh:** File system watcher (fsnotify) monitors the `--path` directory for changes. When a file is created, modified, or deleted, the server pushes an event via SSE (Server-Sent Events). Browser-side JavaScript receives the event and reloads the current page content.

### How it works

1. User runs `kr --path ./docs`
2. CLI parses flags, resolves the directory path, starts the HTTP server on the chosen/random port
3. Server prints `http://localhost:3847` (or whatever port) to stdout
4. User opens the URL in a browser
5. Root URL (`/`) serves the Kanban board: server reads `backlog.md`, parses sections and items, renders the board template
6. Top menu shows folders: server scans the directory for subdirectories, orders known folders first
7. Clicking a folder (e.g., `/features/`) shows a file list: server reads the directory, lists `.md` files
8. Clicking a file (e.g., `/features/2026-03-06-fix-stale-signed-urls.md`) renders the markdown: server reads the file, extracts front matter, converts markdown to HTML, serves the page template
9. Meanwhile, fsnotify watches the directory tree. When an agent updates `backlog.md`, the watcher fires → server broadcasts via SSE → browser reloads the Kanban board automatically

### Visual concept

```
┌──────────────────────────────────────────────────────────────────┐
│  kr   │ Backlog │ Bugs │ Features │ Decisions │ Plans │ ...      │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ Inbox ──┐  ┌─ Ready ──┐  ┌─ Doing ──┐  ┌─ Done ───┐       │
│  │          │  │          │  │          │  │           │        │
│  │ ┌──────┐ │  │ ┌──────┐ │  │ ┌──────┐ │  │ ┌───────┐│        │
│  │ │S-045 │ │  │ │S-043 │ │  │ │S-042 │ │  │ │S-041  ││        │
│  │ │Title │ │  │ │Title │ │  │ │Title │ │  │ │Title  ││        │
│  │ │svc:be│ │  │ │svc:fe│ │  │ │svc:be│ │  │ │svc:be ││        │
│  │ │spec→ │ │  │ │plan→ │ │  │ │plan→ │ │  │ │spec→  ││        │
│  │ └──────┘ │  │ └──────┘ │  │ └──────┘ │  │ └───────┘│        │
│  │          │  │          │  │          │  │           │        │
│  └──────────┘  └──────────┘  └──────────┘  └───────────┘        │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

```
┌──────────────────────────────────────────────────────────────────┐
│  kr   │ Backlog │ Bugs │ Features │ Decisions │ Plans │ ...      │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Features                                                        │
│  ─────────                                                       │
│  📄 2026-03-06-fix-stale-signed-urls.md                         │
│  📄 2026-03-05-scan-upload-flow.md                              │
│  📄 2026-03-04-user-onboarding.md                               │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

```
┌──────────────────────────────────────────────────────────────────┐
│  kr   │ Backlog │ Bugs │ Features │ Decisions │ Plans │ ...      │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ Front Matter ──────────────────────────────────┐            │
│  │ id: S-041  status: done  type: feature          │            │
│  │ service: be  epic: EPIC-003                     │            │
│  └─────────────────────────────────────────────────┘            │
│                                                                  │
│  # Fix Stale Signed URLs in ListScans                           │
│                                                                  │
│  ## Problem                                                      │
│  When listing scans, the signed URLs returned by ...            │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Boundaries

### Explicitly NOT building
- File editing or modification — this is a read-only viewer; agents handle writes
- Search functionality — deferred, can be added as a separate feature later
- Authentication — local tool, no need
- Syntax highlighting for code blocks — deferred, goldmark's raw code blocks are sufficient for now
- Drag-and-drop Kanban — cards are display-only
- Non-markdown file rendering — `.md` files only
- Remote/network file sources — local filesystem only

### Rabbit holes to avoid
- **Over-styling the Kanban** — Cards are functional info displays, not interactive widgets. No animations, no drag-drop, no color-coded statuses. Keep CSS minimal.
- **Full markdown edge cases** — Goldmark handles CommonMark well out of the box. Don't add custom renderers for niche extensions.
- **Deep recursive directory watching** — Watch directories under `--path` one level deep. Don't recursively watch deeply nested structures.
- **Sophisticated debouncing** — Simple debounce (~100ms) for file change events. Don't build a change detection system.
- **Template engine complexity** — Go's `html/template` is sufficient. No need for a custom template language or component system.

## Definition of Done

**The feature is complete when:**

1. `kr --path ./docs` starts a server on a random port and prints the URL to stdout
2. `kr --port 8080 --path ./docs` starts on the specified port
3. Root URL (`/`) renders the Kanban board with backlog items parsed into Inbox, Ready, Doing, Done columns
4. Backlog cards display parsed fields (id, title, service, feature, etc.) and link to referenced docs
5. Top navigation menu shows all folders, with known folders first in fixed order
6. Clicking a folder shows a list of `.md` files in that folder
7. Clicking a `.md` file renders it with proper markdown formatting and front matter displayed
8. Editing a file on disk triggers the browser to auto-refresh within ~1 second (SSE live reload)
9. Missing `backlog.md` shows an empty Kanban board (not an error)
10. Empty folders show a "no documents" message

**Verification:**

Automated:
- [ ] `go test ./...` — all unit tests pass (backlog parser, markdown renderer, route handling)
- [ ] `go build -o kr .` — binary builds without errors
- [ ] `go vet ./...` — no issues

Manual:
- [ ] Start `kr` with a sample docs directory, verify Kanban board renders correctly
- [ ] Click a backlog item link, verify it navigates to the rendered markdown file
- [ ] Navigate through folders via the top menu, verify file listings and document rendering
- [ ] Edit a file in the docs directory, verify the browser auto-refreshes
- [ ] Start with `--port 0` (random), verify URL printed to stdout works
- [ ] Start with `--port 8080`, verify it binds to the specified port

## Success Metrics

**Leading (immediate):**
- Tool serves Kanban board and renders docs correctly on first run
- Live reload works — edit a file, browser updates without manual refresh

**Lagging (2-4 weeks):**
- `kr` is the default way to monitor project docs and agent activity
- No longer opening raw markdown files in terminals

**Failure signal:**
- Still reading raw `.md` files after a week because the viewer is missing something essential

## Implementation Hints

### Key libraries
- `github.com/yuin/goldmark` — markdown to HTML rendering
- `go.abhg.dev/goldmark/frontmatter` — YAML front matter extraction
- `github.com/fsnotify/fsnotify` — cross-platform file system notifications

### Architecture (from stack.md)
```
cmd/            # CLI entry point, flag parsing
internal/
  server/       # HTTP server setup, routing, SSE endpoint
  renderer/     # Goldmark markdown-to-HTML, front matter parsing
  templates/    # Embedded HTML templates (Go embed)
  static/       # Embedded CSS, minimal JS for SSE (Go embed)
```

### Backlog item parsing
Items follow: `- [x] ID: Title | key:value | key:value — status`
- Checkbox: `[x]` = done, `[ ]` = not done
- Fields separated by `|`, each field is `key:value`
- Any value matching `docs/{folder}/{file}.md` is a link
- Trailing text after `—` is a status note

### SSE pattern
- Server endpoint: `GET /events` returns `text/event-stream`
- fsnotify watcher runs in a goroutine, broadcasts to connected SSE clients
- Browser JS: `new EventSource('/events')` listens for messages, calls `location.reload()` or fetches updated content
- Simple debounce: batch rapid file changes into a single SSE event

### Front matter display
- Extract YAML front matter before rendering markdown
- Display as a styled metadata bar/table above the rendered content
- Key-value pairs displayed inline or in a compact grid

### Routing
- `GET /` — Kanban board (backlog.md)
- `GET /{folder}/` — file listing for the folder
- `GET /{folder}/{file}.md` — rendered markdown document
- `GET /events` — SSE endpoint for live reload
- `GET /static/*` — embedded CSS/JS assets

## Research Summary

Light research conducted. Key findings:
- **goldmark** + **goldmark-frontmatter** are the standard Go libraries for markdown rendering with front matter — actively maintained, well-documented
- **fsnotify** is the standard Go file watcher — cross-platform, stable
- **SSE** via stdlib `net/http` requires no external dependencies — simpler than WebSockets for unidirectional push
- No existing tool combines Kanban backlog parsing + markdown rendering + live reload in this way

## Stories

1. **S-001: Set up project skeleton with CLI flags and HTTP server**
   Foundation. Parses `--port` and `--path` flags, starts `net/http` server on chosen/random port, prints URL to stdout. Serves a placeholder page.
   - Acceptance: `kr --port 8080 --path ./docs` starts server, prints URL. Random port works when `--port` not specified.

2. **S-002: Implement markdown rendering with front matter extraction**
   Core rendering engine. Goldmark + frontmatter extension. Takes a `.md` file path, returns rendered HTML + extracted front matter map.
   - Acceptance: Given a markdown file with YAML front matter, produces correct HTML and a map of front matter key-value pairs.
   - Depends on: none

3. **S-003: Create HTML templates and embedded static assets**
   Layout template (nav bar, content area), page templates (backlog, folder list, document view), embedded CSS for clean styling, minimal JS placeholder for SSE.
   - Acceptance: Templates render with proper layout. CSS produces a clean, readable interface. Assets embedded via `go:embed`.
   - Depends on: none

4. **S-004: Implement folder navigation and file listing**
   Scans `--path` for directories, builds nav menu (known folders first, unknown alphabetically). Folder route shows `.md` file list. Document route renders the file.
   - Acceptance: Top menu shows folders in correct order. `/{folder}/` lists files. `/{folder}/{file}.md` renders document with front matter.
   - Depends on: S-002, S-003

5. **S-005: Implement backlog parser and Kanban board view**
   Parses `backlog.md` into sections (Inbox, Ready, Doing, Done). Extracts item fields. Detects `docs/` links. Renders as 4-column Kanban at `/`.
   - Acceptance: Root URL shows Kanban with correct columns. Cards display parsed fields. `docs/` links are clickable. Missing `backlog.md` shows empty board.
   - Depends on: S-003

6. **S-006: Add live reload with fsnotify and SSE**
   File watcher monitors `--path`. SSE endpoint streams change events. Browser JS auto-refreshes. Simple debounce for rapid changes.
   - Acceptance: Edit a file → browser refreshes within ~1 second. Multiple rapid edits = single refresh. SSE reconnects on drop.
   - Depends on: S-004, S-005

## References

- Epic: docs/epics/2026-03-07-live-docs-dashboard.md
- Stack definition: stack.md
- goldmark: https://github.com/yuin/goldmark
- goldmark-frontmatter: https://github.com/abhinav/goldmark-frontmatter
- fsnotify: https://github.com/fsnotify/fsnotify

## Origin

Feature spec created on 2026-03-07 through structured intake from EPIC-001.
Original description: "A Go CLI tool that launches a local web server to render and navigate Markdown documentation files. Includes a Kanban board view for backlog.md, folder-based navigation, markdown rendering, and live auto-refresh when files are updated by AI agents."
