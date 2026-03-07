---
date: 2026-03-07
feature: FEAT-001
spec: docs/features/2026-03-07-live-docs-dashboard.md
status: approved
---

# Implementation Plan: Live Documentation Dashboard

## Overview

We're implementing a Go CLI tool (`kr`) that serves a read-only web dashboard for browsing markdown documentation. The implementation follows a bottom-up approach: foundation (CLI + server), then rendering engine, then templates, then routes/views, and finally live reload. Each phase produces a testable artifact.

This is a greenfield project — all files are new. The folder structure follows `stack.md`: `cmd/` for CLI entry, `internal/server/` for HTTP, `internal/renderer/` for markdown, `internal/templates/` for embedded HTML, `internal/static/` for embedded CSS/JS.

## Reference Implementation

No existing codebase to reference. Patterns drawn from:
- goldmark API: `goldmark.New(WithExtensions(...)).Convert(src, &buf, parser.WithContext(ctx))`
- goldmark-frontmatter: `frontmatter.Get(ctx).Decode(&meta)` to extract YAML front matter into `map[string]any`
- fsnotify: `watcher.Add(path)` per directory, listen on `watcher.Events` channel, `event.Has(fsnotify.Write)`
- SSE: stdlib `net/http` with `text/event-stream` content type, `Flusher` interface

## Pre-conditions

Before starting implementation:
- [x] Feature spec is approved: docs/features/2026-03-07-live-docs-dashboard.md
- [ ] Go 1.25.0 installed and available

---

## Phase 1: Project Foundation — CLI & HTTP Server

### Overview
Set up the Go project structure, add dependencies, parse CLI flags, and start an HTTP server that serves a placeholder page. This is the skeleton everything else builds on.

### Step 1.1: Add Dependencies
**File:** `go.mod` (modify)

**What to do:**
Run `go get` to add the three external dependencies:
- `github.com/yuin/goldmark`
- `go.abhg.dev/goldmark/frontmatter`
- `github.com/fsnotify/fsnotify`

This updates `go.mod` and creates `go.sum`.

### Step 1.2: Create CLI Entry Point
**File:** `cmd/kr/main.go` (create)

**What to do:**
Create the `main` package with:
- `flag.Int` for `--port` (default `0` for random)
- `flag.String` for `--path` (default `.`)
- Validate that `--path` exists and is a directory
- Resolve `--path` to an absolute path
- Call `server.Start(port, absPath)` to start the HTTP server
- Print `http://localhost:<actualPort>` to stdout after the server binds

The main function should:
1. Parse flags
2. Validate the path (stat it, check `IsDir()`)
3. Create the server and start it
4. Block until interrupted (signal handling with `os/signal`)

### Step 1.3: Create Server Package
**File:** `internal/server/server.go` (create)

**What to do:**
Create a `Server` struct that holds:
- `port int` — the requested port (0 = random)
- `rootPath string` — absolute path to the docs directory
- `mux *http.ServeMux` — the router

Implement a `New(port int, rootPath string) *Server` constructor that:
- Creates a new `http.ServeMux`
- Registers a temporary placeholder route on `/` that returns `"kr is running"`

Implement a `Start() error` method that:
- Creates a `net.Listener` on `tcp`, `:<port>` (port 0 lets the OS pick)
- Extracts the actual port from `listener.Addr().(*net.TCPAddr).Port`
- Prints `http://localhost:<port>` to stdout via `fmt.Println`
- Calls `http.Serve(listener, s.mux)`

### Step 1.4: Wire Main to Server
**File:** `cmd/kr/main.go` (modify — complete the wiring)

**What to do:**
In `main()`, after flag parsing and validation:
- Create `srv := server.New(port, absPath)`
- Call `srv.Start()` in a goroutine
- Wait for `os.Interrupt` signal to shut down gracefully

### Phase 1 Verification

**Automated:**
- [ ] `go build -o kr ./cmd/kr` — binary builds without errors
- [ ] `go vet ./...` — no issues

**Manual:**
- [ ] `./kr --path ./docs` starts and prints a URL with a random port
- [ ] `./kr --port 8080 --path ./docs` starts on port 8080
- [ ] Visiting the URL in a browser shows "kr is running"
- [ ] `./kr --path /nonexistent` exits with an error message

**Stop here.** Verify Phase 1 before proceeding.

---

## Phase 2: Markdown Rendering Engine

### Overview
Build the rendering package that converts markdown files to HTML with front matter extraction. This is a pure library with no HTTP dependencies — it takes bytes in and returns structured output.

### Step 2.1: Create Renderer Package
**File:** `internal/renderer/renderer.go` (create)

**What to do:**
Create a `Renderer` struct that holds a configured `goldmark.Markdown` instance.

Implement `New() *Renderer` that creates a goldmark instance with:
- `goldmark.WithExtensions(&frontmatter.Extender{})` — front matter support
- `goldmark.WithExtensions(extension.GFM)` — GitHub Flavored Markdown (tables, strikethrough, autolinks)
- `goldmark.WithParserOptions(parser.WithAutoHeadingID())` — anchor IDs on headings
- `goldmark.WithRendererOptions(html.WithUnsafe())` — allow raw HTML in markdown (needed for some docs)

Define a `Result` struct:
```go
type Result struct {
    HTML        string
    FrontMatter map[string]any
}
```

Implement `Render(source []byte) (*Result, error)`:
1. Create a `parser.NewContext()`
2. Call `r.md.Convert(source, &buf, parser.WithContext(ctx))`
3. Get front matter: `d := frontmatter.Get(ctx)` — if `d != nil`, decode into `map[string]any`
4. Return `&Result{HTML: buf.String(), FrontMatter: meta}`

### Step 2.2: Create Renderer Tests
**File:** `internal/renderer/renderer_test.go` (create)

**What to do:**
Write tests for:
- **Basic markdown:** headings, paragraphs, lists, links, code blocks → verify HTML output contains expected tags
- **Front matter extraction:** markdown with `---` delimited YAML → verify `FrontMatter` map contains correct keys and values
- **No front matter:** markdown without front matter → verify `FrontMatter` is nil or empty, HTML renders correctly
- **GFM tables:** markdown with pipe tables → verify `<table>` rendered
- **Empty input:** empty bytes → verify no error, empty result

### Phase 2 Verification

**Automated:**
- [ ] `go test ./internal/renderer/...` — all tests pass
- [ ] `go vet ./...` — no issues

**Stop here.** Verify Phase 2 before proceeding.

---

## Phase 3: HTML Templates & Embedded Static Assets

### Overview
Create the HTML templates (layout, backlog, folder list, document view) and static assets (CSS, JS) that will be embedded into the binary. After this phase, the templates exist and can be parsed — but they're not yet wired to routes.

### Step 3.1: Create Layout Template
**File:** `internal/templates/layout.html` (create)

**What to do:**
Create the base HTML layout that all pages extend. It should include:
- Standard HTML5 doctype, `<head>` with charset, viewport meta, title (`kr`)
- `<link>` to `/static/style.css`
- Navigation bar: a `<nav>` element with:
  - App name "kr" linking to `/` (home/backlog)
  - Placeholder `{{.NavItems}}` block for folder links — each is an `<a>` with the folder name, linking to `/{folder}/`
  - Active state class for the current folder
- Content area: `{{.Content}}` block for page-specific content
- `<script src="/static/main.js"></script>` at the end of body

Use Go's `html/template` with `{{define "layout"}}...{{end}}` and `{{template "content" .}}` for content injection.

### Step 3.2: Create Backlog Template
**File:** `internal/templates/backlog.html` (create)

**What to do:**
Create the Kanban board template that renders inside the layout's content area. Structure:
- A container `<div class="kanban">` with CSS flexbox for column layout
- Four columns, each a `<div class="kanban-column">`:
  - Column header: `<h2>` with section name and item count badge
  - Card list: iterate over items in each section
- Each card is a `<div class="kanban-card">`:
  - Card title: the item ID and title text (e.g., "S-041: Generate fresh signed URLs...")
  - Checkbox indicator: visual done/not-done state
  - Fields area: iterate over parsed key-value fields, render each as a label-value pair
  - For fields whose values are links (contain `docs/`), render as `<a href="/{path}">` within the viewer
  - Status note at the bottom if present (the text after `—`)

Template data structure to expect:
```go
type BacklogData struct {
    Sections []Section  // ordered: Inbox, Ready, Doing, Done
    NavItems []NavItem
}
type Section struct {
    Name  string
    Items []BacklogItem
}
type BacklogItem struct {
    Done    bool
    ID      string
    Title   string
    Fields  []Field
    Status  string  // text after —
}
type Field struct {
    Key   string
    Value string
    IsLink bool
    LinkPath string  // internal path for the viewer
}
```

### Step 3.3: Create Folder List Template
**File:** `internal/templates/folder.html` (create)

**What to do:**
Create the file listing template for when a user navigates to a folder. Structure:
- Folder name as `<h1>`
- A list of `.md` files as `<ul>`:
  - Each file is an `<li>` with an `<a>` linking to `/{folder}/{filename}`
  - Display the filename (without path prefix)
- If no files, show a "No documents in this folder" message

### Step 3.4: Create Document View Template
**File:** `internal/templates/document.html` (create)

**What to do:**
Create the markdown document view template. Structure:
- If front matter exists, render a `<div class="front-matter">` section:
  - Display as a compact grid/table of key-value pairs
  - Each pair: `<span class="fm-key">key:</span> <span class="fm-value">value</span>`
- Render the HTML content in a `<div class="markdown-body">` using `{{.Content}}` (pre-rendered HTML from goldmark, so use `template.HTML` type to avoid escaping)

### Step 3.5: Create Template Loader
**File:** `internal/templates/templates.go` (create)

**What to do:**
Use `//go:embed *.html` to embed all template files. Create a `Templates` struct with:
- A parsed `*template.Template` tree that includes all templates
- Methods to render each page type:
  - `RenderBacklog(w io.Writer, data BacklogData) error`
  - `RenderFolder(w io.Writer, data FolderData) error`
  - `RenderDocument(w io.Writer, data DocumentData) error`

Each render method executes the layout template with the appropriate content template nested inside.

Define the data types (`BacklogData`, `FolderData`, `DocumentData`, `NavItem`, `Section`, `BacklogItem`, `Field`) in this package.

### Step 3.6: Create CSS
**File:** `internal/static/style.css` (create)

**What to do:**
Create a clean, minimal stylesheet. Key rules:
- **Reset/base:** box-sizing, system font stack, reasonable line-height, muted background
- **Navigation bar:** fixed top, horizontal flex, dark background, white text, folder links as pills/tabs, active state highlighted
- **Layout:** content area below nav, max-width ~900px centered, padding
- **Kanban board:** `.kanban` with `display: flex; gap: 1rem; overflow-x: auto`. Columns flex equally. Column headers bold with count badges.
- **Kanban cards:** `.kanban-card` with white background, subtle border/shadow, padding, margin-bottom. Done cards slightly muted (opacity or strikethrough). Fields as small label-value pairs. Links in blue.
- **Folder listing:** simple `<ul>` with file links, clean spacing
- **Document view:** `.markdown-body` with readable typography — headings sized down from h1, code blocks with background, blockquotes with left border, tables with borders, links in blue
- **Front matter bar:** `.front-matter` with subtle background, compact key-value grid, small font size, rounded corners
- **Responsive:** single column kanban on narrow screens (stack columns vertically)

Keep it under ~200 lines. No CSS framework — vanilla CSS.

### Step 3.7: Create JS Placeholder
**File:** `internal/static/main.js` (create)

**What to do:**
Create a minimal JS file with a placeholder comment for SSE live reload (will be implemented in Phase 5). For now, just:
```js
// Live reload via SSE — implemented in Phase 5
console.log('kr loaded');
```

### Step 3.8: Create Static Asset Embedder
**File:** `internal/static/static.go` (create)

**What to do:**
Use `//go:embed style.css main.js` to embed static files. Export an `http.FileServer` handler or the `embed.FS` variable that the server can mount at `/static/`.

### Phase 3 Verification

**Automated:**
- [ ] `go build ./...` — all packages compile (templates parse, embeds resolve)
- [ ] `go vet ./...` — no issues

**Manual:**
- [ ] Template files exist and contain valid Go template syntax
- [ ] CSS file is readable and has rules for all specified selectors

**Stop here.** Verify Phase 3 before proceeding.

---

## Phase 4: Routes — Navigation, File Listing, Document View & Kanban Board

### Overview
Wire everything together: the server routes, folder scanning, backlog parsing, and Kanban rendering. After this phase, all pages work — you can browse folders, read documents, and see the Kanban board. Only live reload is missing.

### Step 4.1: Create Backlog Parser
**File:** `internal/backlog/parser.go` (create)

**What to do:**
Create a `backlog` package that parses `backlog.md` into structured data.

Implement `Parse(content []byte) (*Backlog, error)`:
1. Split content by lines
2. Track current section by detecting `## SectionName` headers
3. For each line starting with `- [x]` or `- [ ]`:
   - Extract checkbox state: `[x]` = done, `[ ]` = not done
   - Extract the text after the checkbox
   - Split on ` — ` (em dash with spaces) to separate the main content from the status note
   - From the main content, extract the ID and title: everything before the first `|` is `ID: Title`
   - Parse `ID: Title` by splitting on `: ` — first part is ID, rest is title
   - Split the remaining text on `|` to get fields
   - Each field is `key:value` — trim spaces, split on first `:`
   - For each field value, check if it contains `docs/` — if so, mark as a link and extract the path (strip `docs/` prefix to make the internal viewer path `/{folder}/{file}`)
4. Order sections as: Inbox, Ready, Doing, Done (regardless of order in the file)
5. If a section doesn't exist in the file, include it as empty

Define types:
```go
type Backlog struct {
    Sections []Section
}
type Section struct {
    Name  string
    Items []Item
}
type Item struct {
    Done   bool
    ID     string
    Title  string
    Fields []Field
    Status string
}
type Field struct {
    Key      string
    Value    string
    IsLink   bool
    LinkPath string
}
```

### Step 4.2: Create Backlog Parser Tests
**File:** `internal/backlog/parser_test.go` (create)

**What to do:**
Write tests for:
- **Full backlog parsing:** A sample backlog with all 4 sections, multiple items → verify correct section ordering, item count, field parsing
- **Item field extraction:** `- [x] S-041: Title | feature:fix | service:be | plan:docs/plans/file.md` → verify ID, title, fields, link detection
- **Checkbox state:** `[x]` → done=true, `[ ]` → done=false
- **Status note:** text after `—` extracted correctly
- **Missing sections:** Backlog with only `## Ready` and `## Done` → all 4 sections present, missing ones empty
- **Empty backlog:** Just headers, no items → sections exist but are empty
- **Link detection:** `plan:docs/plans/file.md` → IsLink=true, LinkPath="/plans/file.md"; `service:be` → IsLink=false

### Step 4.3: Create Folder Scanner
**File:** `internal/server/folders.go` (create)

**What to do:**
Implement folder scanning logic for navigation and file listing.

Define known folders with their display order:
```go
var knownFolders = []string{"bugs", "features", "decisions", "plans", "research", "reviews", "handoffs"}
```

Implement `scanFolders(rootPath string) ([]NavItem, error)`:
1. Read the root directory entries with `os.ReadDir(rootPath)`
2. Filter to directories only (skip files, skip hidden dirs starting with `.`)
3. Separate into known folders (preserve fixed order) and unknown folders (sort alphabetically)
4. Combine: known first, then unknown
5. Return as `[]NavItem` with `Name` and `Path` fields

Implement `listFiles(folderPath string) ([]FileEntry, error)`:
1. Read the directory entries
2. Filter to `.md` files only
3. Sort by filename descending (newest date-prefixed files first)
4. Return as `[]FileEntry` with `Name` and `Path`

### Step 4.4: Register All Routes
**File:** `internal/server/server.go` (modify)

**What to do:**
Replace the placeholder route with the full routing setup. In the `New()` constructor or a `setupRoutes()` method:

- `GET /` — Kanban board handler
- `GET /static/` — serve embedded static assets via `http.FileServer(http.FS(static.FS))` with `http.StripPrefix("/static/", ...)`
- `GET /{folder}/` — folder file listing handler
- `GET /{folder}/{file}` — document view handler

Go 1.22+ supports path parameters in `http.ServeMux` patterns. Use:
- `mux.HandleFunc("GET /{folder}/", s.handleFolder)`
- `mux.HandleFunc("GET /{folder}/{file}", s.handleDocument)`

Store the `renderer.Renderer`, `templates.Templates`, and `rootPath` on the `Server` struct so handlers can access them.

### Step 4.5: Implement Kanban Board Handler
**File:** `internal/server/handlers.go` (create)

**What to do:**
Implement `handleBacklog(w http.ResponseWriter, r *http.Request)`:
1. Read `backlog.md` from `rootPath` — if file doesn't exist, render empty board (not an error)
2. Parse with `backlog.Parse(content)`
3. Scan folders for nav items
4. Construct `BacklogData` with sections and nav items
5. Call `templates.RenderBacklog(w, data)`

### Step 4.6: Implement Folder Handler
**File:** `internal/server/handlers.go` (modify — add to same file)

**What to do:**
Implement `handleFolder(w http.ResponseWriter, r *http.Request)`:
1. Extract `{folder}` from the request path using `r.PathValue("folder")`
2. Construct the folder path: `filepath.Join(rootPath, folder)`
3. Validate the folder exists and is a directory — if not, return 404
4. List `.md` files in the folder
5. Scan folders for nav items (to render the nav bar with active state)
6. Construct `FolderData` and render with `templates.RenderFolder(w, data)`

### Step 4.7: Implement Document Handler
**File:** `internal/server/handlers.go` (modify — add to same file)

**What to do:**
Implement `handleDocument(w http.ResponseWriter, r *http.Request)`:
1. Extract `{folder}` and `{file}` from the request path
2. Validate `{file}` ends with `.md` — if not, return 404
3. Construct file path: `filepath.Join(rootPath, folder, file)`
4. Read the file — if not found, return 404
5. Render with `renderer.Render(content)`
6. Scan folders for nav items
7. Construct `DocumentData` with rendered HTML, front matter, nav items, current folder/file
8. Render with `templates.RenderDocument(w, data)`

**Security:** Validate that the resolved file path is still within `rootPath` (prevent path traversal via `../`). Use `filepath.Rel` or check that `filepath.Clean(path)` starts with `rootPath`.

### Step 4.8: Handler Tests
**File:** `internal/server/handlers_test.go` (create)

**What to do:**
Write HTTP handler tests using `httptest.NewServer` or `httptest.NewRecorder`:
- **Root URL:** GET `/` → 200, response contains "kanban" class
- **Folder listing:** Create a temp dir with `.md` files, GET `/{folder}/` → 200, response contains file names
- **Document view:** Create a temp `.md` file with front matter, GET `/{folder}/{file}.md` → 200, response contains rendered HTML
- **Missing folder:** GET `/nonexistent/` → 404
- **Missing file:** GET `/features/nonexistent.md` → 404
- **Path traversal:** GET `/../../etc/passwd` → 404 (not 200)
- **Missing backlog:** No `backlog.md` in root → GET `/` returns 200 with empty board

### Step 4.9: Backlog Parser Tests for Link Detection
**File:** `internal/backlog/parser_test.go` (modify — add test cases)

**What to do:**
Add specific test cases for the link path transformation:
- `plan:docs/plans/2026-03-06-file.md` → LinkPath should be `/plans/2026-03-06-file.md`
- `spec:docs/features/2026-03-06-file.md` → LinkPath should be `/features/2026-03-06-file.md`
- `epic:docs/epics/2026-03-07-epic.md` → LinkPath should be `/epics/2026-03-07-epic.md`
- `service:be` → not a link
- Values without `docs/` prefix → not links

### Phase 4 Verification

**Automated:**
- [ ] `go test ./...` — all tests pass (renderer, backlog parser, handlers)
- [ ] `go build -o kr ./cmd/kr` — binary builds
- [ ] `go vet ./...` — no issues

**Manual:**
- [ ] Start `kr` with the project's own `docs/` directory
- [ ] Root URL shows the Kanban board with the 6 stories from the backlog
- [ ] Story cards show parsed fields (feature, spec, depends)
- [ ] Spec links on cards navigate to the feature spec rendered as HTML
- [ ] Top nav menu shows: bugs, features, decisions, plans, research, reviews, handoffs (known folders in order)
- [ ] Clicking "features" shows the feature spec file
- [ ] Clicking the feature spec renders the markdown with front matter displayed
- [ ] Visiting a non-existent folder returns 404
- [ ] Visiting a non-existent file returns 404

**Stop here.** Verify Phase 4 before proceeding.

---

## Phase 5: Live Reload — fsnotify & SSE

### Overview
Add file system watching and Server-Sent Events so the browser auto-refreshes when files change. This is the final phase that completes the feature.

### Step 5.1: Create File Watcher
**File:** `internal/watcher/watcher.go` (create)

**What to do:**
Create a `Watcher` struct that wraps fsnotify and provides a channel of change events.

Implement `New(rootPath string) (*Watcher, error)`:
1. Create `fsnotify.NewWatcher()`
2. Walk the `rootPath` directory one level deep — add the root directory and each immediate subdirectory to the watcher (fsnotify is non-recursive)
3. Start a goroutine that listens on `watcher.Events`:
   - On any Create, Write, Remove, or Rename event for `.md` files:
     - Debounce: use a `time.Timer` (100ms). Reset the timer on each event. When the timer fires, send a notification on an output channel.
   - On Create events for directories: add the new directory to the watcher (handles new folders appearing)
4. Return the `Watcher` which exposes a `Changes() <-chan struct{}` channel

Implement `Close() error` to clean up the fsnotify watcher.

### Step 5.2: Create SSE Endpoint
**File:** `internal/server/sse.go` (create)

**What to do:**
Create an SSE broker that manages connected clients and broadcasts change events.

Implement an `SSEBroker` struct with:
- `clients map[chan struct{}]struct{}` — set of connected client channels
- `mu sync.Mutex` for thread-safe client management
- Methods: `Subscribe() chan struct{}`, `Unsubscribe(ch chan struct{})`, `Broadcast()`

Implement `handleSSE(w http.ResponseWriter, r *http.Request)`:
1. Set headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`
2. Check that `w` implements `http.Flusher` — if not, return 500
3. Subscribe to the broker
4. Defer unsubscribe
5. Send an initial `event: connected\ndata: ok\n\n` and flush
6. Loop:
   - `select` on the client channel and `r.Context().Done()`
   - On client channel receive: write `event: change\ndata: reload\n\n` and flush
   - On context done: return (client disconnected)

### Step 5.3: Wire Watcher to SSE Broker
**File:** `internal/server/server.go` (modify)

**What to do:**
In the `Server` struct, add:
- `watcher *watcher.Watcher`
- `broker *SSEBroker`

In `New()`:
- Create the watcher: `watcher.New(rootPath)`
- Create the SSE broker
- Start a goroutine that reads from `watcher.Changes()` and calls `broker.Broadcast()`
- Register the SSE route: `mux.HandleFunc("GET /events", s.handleSSE)`

In `Start()` or a cleanup method:
- Defer `watcher.Close()` on shutdown

### Step 5.4: Implement Browser-Side SSE Client
**File:** `internal/static/main.js` (modify — replace placeholder)

**What to do:**
Replace the placeholder with the SSE client:
```js
(function() {
  var source = new EventSource('/events');

  source.addEventListener('change', function() {
    location.reload();
  });

  source.addEventListener('error', function() {
    // SSE connection lost — will auto-reconnect (browser default behavior)
    console.log('SSE connection lost, reconnecting...');
  });

  source.addEventListener('connected', function() {
    console.log('Live reload connected');
  });
})();
```

The browser's `EventSource` automatically reconnects on connection loss with exponential backoff — no custom reconnection logic needed.

### Step 5.5: Watcher Tests
**File:** `internal/watcher/watcher_test.go` (create)

**What to do:**
Write tests for:
- **File modification:** Create a temp dir with an `.md` file, start watcher, modify the file → receive event on `Changes()` channel within 500ms
- **File creation:** Create a new `.md` file in a watched dir → receive event
- **Debouncing:** Rapidly modify a file 5 times within 50ms → receive only 1 event (after the 100ms debounce window)
- **Non-md files ignored:** Create/modify a `.txt` file → no event (only `.md` changes trigger reload)
- **Subdirectory watching:** Create an `.md` file in a subdirectory → receive event

### Step 5.6: SSE Integration Test
**File:** `internal/server/sse_test.go` (create)

**What to do:**
Write tests for:
- **SSE connection:** GET `/events` → response has `Content-Type: text/event-stream`, receives `connected` event
- **Broadcast:** Connect to SSE, trigger a broadcast on the broker → client receives `change` event
- **Multiple clients:** Connect 2 clients, broadcast → both receive the event
- **Client disconnect:** Connect, then close the connection → client is unsubscribed (no goroutine leak)

### Phase 5 Verification

**Automated:**
- [ ] `go test ./...` — all tests pass (including watcher and SSE tests)
- [ ] `go build -o kr ./cmd/kr` — binary builds
- [ ] `go vet ./...` — no issues

**Manual:**
- [ ] Start `kr` with `docs/` directory
- [ ] Open the browser, check browser console shows "Live reload connected"
- [ ] Edit `docs/backlog.md` in a text editor → browser auto-refreshes within ~1 second
- [ ] Navigate to a document page, edit that document → browser auto-refreshes
- [ ] Close and reopen the browser tab → SSE reconnects automatically
- [ ] Create a new `.md` file in a folder → refreshes; folder listing includes the new file

**Stop here.** Verify Phase 5 before proceeding to final verification.

---

## Final Verification

**All automated checks:**
- [ ] Full test suite passes: `go test ./...`
- [ ] Linting passes: `go vet ./...`
- [ ] Build succeeds: `go build -o kr ./cmd/kr`

**Manual testing — end to end:**
- [ ] `./kr --path ./docs` — starts on random port, URL printed, opens in browser
- [ ] `./kr --port 8080 --path ./docs` — starts on port 8080
- [ ] Root URL shows Kanban board with correct columns (Inbox, Ready, Doing, Done)
- [ ] Backlog cards show parsed fields with clickable links to docs
- [ ] Top nav shows known folders in order, then any unknown folders
- [ ] Click a folder → see list of `.md` files
- [ ] Click a file → see rendered markdown with front matter metadata
- [ ] Edit any file on disk → browser auto-refreshes within ~1 second
- [ ] Missing `backlog.md` → empty Kanban board (no error)
- [ ] Empty folder → "No documents" message
- [ ] `./kr --path /nonexistent` → error message, exits

**Definition of done alignment:**
- [ ] DoD 1 (random port startup) — addressed in Phase 1, Step 1.2-1.3
- [ ] DoD 2 (specific port) — addressed in Phase 1, Step 1.2-1.3
- [ ] DoD 3 (Kanban board with columns) — addressed in Phase 4, Step 4.1-4.5
- [ ] DoD 4 (cards with parsed fields and links) — addressed in Phase 4, Step 4.1-4.5
- [ ] DoD 5 (nav menu with known folders first) — addressed in Phase 4, Step 4.3-4.4
- [ ] DoD 6 (folder file listing) — addressed in Phase 4, Step 4.6
- [ ] DoD 7 (rendered markdown with front matter) — addressed in Phase 2 + Phase 4, Step 4.7
- [ ] DoD 8 (live reload within ~1 second) — addressed in Phase 5
- [ ] DoD 9 (missing backlog → empty board) — addressed in Phase 4, Step 4.5
- [ ] DoD 10 (empty folders → message) — addressed in Phase 4, Step 4.6

## Files Changed Summary

| File | Action | Phase | Notes |
|---|---|---|---|
| `go.mod` | modify | 1 | Add goldmark, goldmark-frontmatter, fsnotify dependencies |
| `go.sum` | create | 1 | Auto-generated by `go get` |
| `cmd/kr/main.go` | create | 1 | CLI entry point, flag parsing, signal handling |
| `internal/server/server.go` | create | 1 | HTTP server setup, mux, listener, Start() |
| `internal/renderer/renderer.go` | create | 2 | Goldmark renderer with front matter extraction |
| `internal/renderer/renderer_test.go` | create | 2 | Renderer unit tests |
| `internal/templates/layout.html` | create | 3 | Base HTML layout with nav bar |
| `internal/templates/backlog.html` | create | 3 | Kanban board template |
| `internal/templates/folder.html` | create | 3 | Folder file listing template |
| `internal/templates/document.html` | create | 3 | Markdown document view template |
| `internal/templates/templates.go` | create | 3 | Template loader, embed, render methods, data types |
| `internal/static/style.css` | create | 3 | Stylesheet for all views |
| `internal/static/main.js` | create | 3 | JS placeholder (SSE client added in Phase 5) |
| `internal/static/static.go` | create | 3 | Static asset embedder |
| `internal/backlog/parser.go` | create | 4 | Backlog markdown parser |
| `internal/backlog/parser_test.go` | create | 4 | Parser unit tests |
| `internal/server/folders.go` | create | 4 | Folder scanning and file listing |
| `internal/server/handlers.go` | create | 4 | HTTP handlers for all routes |
| `internal/server/handlers_test.go` | create | 4 | Handler integration tests |
| `internal/watcher/watcher.go` | create | 5 | fsnotify wrapper with debouncing |
| `internal/watcher/watcher_test.go` | create | 5 | Watcher unit tests |
| `internal/server/sse.go` | create | 5 | SSE broker and endpoint handler |
| `internal/server/sse_test.go` | create | 5 | SSE integration tests |
| `internal/static/main.js` | modify | 5 | Add SSE client for live reload |

**Total: 1 modified, 22 created**

## Risks and Fallbacks

- **fsnotify platform quirks:** macOS has a limit on watched file descriptors. Fallback: if the docs directory has many subdirectories, only watch the top-level + known folders, skip deeply nested dirs.
- **SSE connection limits:** Browsers limit concurrent SSE connections per domain (~6). Since this is localhost with typically 1 tab, this is unlikely to be an issue. Fallback: if multiple tabs are needed, SSE will queue — acceptable for a local tool.
- **Template parsing errors at startup:** If templates have syntax errors, the binary will panic on startup. Mitigation: validate templates in tests (Phase 3 verification).
- **Large markdown files:** Very large `.md` files could be slow to render. Fallback: goldmark is fast (comparable to cmark C implementation) — unlikely to be an issue for documentation files.

## References

- Feature spec: docs/features/2026-03-07-live-docs-dashboard.md
- Epic: docs/epics/2026-03-07-live-docs-dashboard.md
- Stack definition: stack.md
- goldmark API: https://pkg.go.dev/github.com/yuin/goldmark
- goldmark-frontmatter API: https://pkg.go.dev/go.abhg.dev/goldmark/frontmatter
- fsnotify API: https://pkg.go.dev/github.com/fsnotify/fsnotify
