---
id: EPIC-001
date: 2026-03-07
status: active
po_recommendation: strong-yes
affected_repos: [kr]
decisions: []
---

# Live Documentation Dashboard

## Product Vision

> **One sentence:** A local web server that renders a project's markdown documentation as a live, auto-refreshing dashboard with Kanban backlog view — the control panel for AI-driven development workflows.

## Problem Statement

When running AI agents that continuously update project documentation (backlog, features, plans, bugs, decisions), there is no convenient way to monitor their activity in real time. The current experience involves reading raw markdown files in a terminal, manually refreshing, and mentally assembling the project state from scattered files.

This is especially painful for the backlog — a structured file with Inbox, Ready, Doing, and Done sections that maps naturally to a Kanban board but is consumed as raw text. Links between items (specs, plans, epics) require manual navigation between files.

The need is immediate: the more AI agents you run, the harder it becomes to maintain situational awareness without a dedicated viewer.

## User Impact

**Who benefits:** Solo developer using AI-agent workflows to manage project documentation.

**Current experience:** Reading raw `.md` files via `cat`, editors, or terminal tools. Manually navigating between referenced documents. No live updates — must re-read files to see agent changes. No visual backlog overview.

**Target experience:** Run `kr --path ./docs`, open a browser, and see:
- A Kanban board showing the backlog with items as cards, fields parsed, and links clickable
- A top navigation menu listing all documentation folders
- Rendered markdown with proper formatting and front matter display
- Live auto-refresh — when agents update files, the browser updates immediately

## Product Analysis Summary

**Market context:** Existing markdown viewers (Grip, Glow, mdBook, MkDocs) are either terminal-based, require build steps, or lack the Kanban/live-reload combination. None are purpose-built for the AI-agent documentation workflow pattern.

**Competitive landscape:** No direct competitor in this niche. The closest tools are generic markdown servers that lack backlog parsing and live reload.

**Recommendation:** STRONG YES — clear pain point, no existing solution, low build cost, high daily leverage.

**Biggest risk:** Scope creep — adding editing, search, or interactive features. The read-only constraint is key to shipping fast.

## Success Metrics

**Leading indicators:**
- Tool is usable within the first sprint — can view backlog as Kanban and browse folders
- Replaces terminal-based file reading for monitoring agent activity

**Lagging indicators:**
- Becomes the default way to monitor AI agent activity across all projects
- Used daily as the primary docs viewer

**Failure signal:**
- Still reading raw markdown files in the terminal after a week of having `kr` available

## Affected Repos

| Repo | Role | Work Summary |
|---|---|---|
| kr | Sole repo | All implementation — CLI, server, renderer, templates, file watcher |

## Cross-Team Agreements

None — standalone project.

## Scope

### In scope
- CLI with `--port` (optional, random by default) and `--path` (docs directory) flags
- HTTP server printing the port to console on startup
- Kanban board view for `backlog.md` at the root URL (`/`)
- Backlog parsing: sections (Inbox, Ready, Doing, Done) displayed in that order
- Backlog item parsing: extract fields (id, title, feature, service, depends, plan, spec, epic, etc.) and display as card content
- Backlog item links: any field value matching `docs/{folder}/{file}.md` becomes a clickable link to the rendered page
- Top navigation menu from folders in the target directory
- Known folders (bugs, features, decisions, plans, research, reviews, handoffs) shown with priority; unknown folders also included
- Folder view: list of markdown files when clicking a folder (`/{folder}/`)
- Document view: rendered markdown when clicking a file (`/{folder}/{file}.md`)
- Markdown rendering via goldmark with proper formatting (headings, lists, links, code blocks, etc.)
- Front matter extraction and display
- File system watching with live auto-refresh via SSE (Server-Sent Events)
- Embedded templates and static assets via Go `embed`
- Clean, readable web interface

### Out of scope (explicitly)
- Editing or modifying files (read-only tool)
- Search functionality
- Authentication or multi-user support
- Syntax highlighting for code blocks (can be added later)
- PDF or non-markdown file rendering
- Remote file sources (local filesystem only)
- Plugin system or extensibility

### Open Questions
- [ ] Should the Kanban columns be collapsible?
- [ ] Should there be a count badge on each column header?
- [ ] Should the navigation highlight the currently active folder/file?
- [ ] What styling approach — minimal custom CSS or a lightweight CSS framework (e.g., classless CSS)?

## Next Steps

1. Run `/feature` in this repo to create the technical feature spec from this epic
2. Run `/plan` to break the feature into implementable phases
3. Implement phase by phase

## Origin

Epic created on 2026-03-07 through structured intake.
Original description: "A Go CLI tool that launches a local web server to render and navigate Markdown documentation files. Includes a Kanban board view for backlog.md, folder-based navigation, markdown rendering, and live auto-refresh when files are updated by AI agents."
