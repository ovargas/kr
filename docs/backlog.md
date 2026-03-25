# Backlog

## Doing
<!-- Items currently in progress -->
- [=] S-010: Show project folder name in nav bar and browser tab title | feature:project-name-in-navbar | plan:docs/plans/2026-03-13-project-name-in-navbar.md | spec:docs/features/2026-03-13-project-name-in-navbar.md
- [=] S-011: Implement search logic — regex compilation and file scanning | feature:content-search | service:be | plan:docs/plans/2026-03-25-content-search.md | spec:docs/features/2026-03-25-content-search.md
- [=] S-012: Add search handler and route | feature:content-search | service:be | depends:S-011 | plan:docs/plans/2026-03-25-content-search.md | spec:docs/features/2026-03-25-content-search.md
- [=] S-013: Add search results template | feature:content-search | service:fe | depends:S-012 | plan:docs/plans/2026-03-25-content-search.md | spec:docs/features/2026-03-25-content-search.md
- [=] S-014: Add search input and toggles to navbar | feature:content-search | service:fe | depends:S-013 | plan:docs/plans/2026-03-25-content-search.md | spec:docs/features/2026-03-25-content-search.md

## Ready
<!-- Items ready for implementation — refined and planned -->

## Done
<!-- Completed items -->
- [x] BUG: Excerpt shows content before first section, not section content | severity:medium | bug:docs/bugs/2026-03-07-excerpt-wrong-section.md
- [x] S-009: Update folder template for rich file listing | feature:rich-file-listing | depends:S-008 | plan:docs/plans/2026-03-07-rich-file-listing.md | spec:docs/features/2026-03-07-rich-file-listing.md
- [x] S-008: Extend file listing to include title and excerpt | feature:rich-file-listing | depends:S-007 | plan:docs/plans/2026-03-07-rich-file-listing.md | spec:docs/features/2026-03-07-rich-file-listing.md
- [x] S-007: Add title and excerpt extraction for markdown files | feature:rich-file-listing | plan:docs/plans/2026-03-07-rich-file-listing.md | spec:docs/features/2026-03-07-rich-file-listing.md
- [x] S-001: Set up project skeleton with CLI flags and HTTP server | feature:live-docs-dashboard | plan:docs/plans/2026-03-07-live-docs-dashboard.md | spec:docs/features/2026-03-07-live-docs-dashboard.md
- [x] S-002: Implement markdown rendering with front matter extraction | feature:live-docs-dashboard | plan:docs/plans/2026-03-07-live-docs-dashboard.md | spec:docs/features/2026-03-07-live-docs-dashboard.md
- [x] S-003: Create HTML templates and embedded static assets | feature:live-docs-dashboard | plan:docs/plans/2026-03-07-live-docs-dashboard.md | spec:docs/features/2026-03-07-live-docs-dashboard.md
- [x] S-004: Implement folder navigation and file listing | feature:live-docs-dashboard | depends:S-002,S-003 | plan:docs/plans/2026-03-07-live-docs-dashboard.md | spec:docs/features/2026-03-07-live-docs-dashboard.md
- [x] S-005: Implement backlog parser and Kanban board view | feature:live-docs-dashboard | depends:S-003 | plan:docs/plans/2026-03-07-live-docs-dashboard.md | spec:docs/features/2026-03-07-live-docs-dashboard.md
- [x] S-006: Add live reload with fsnotify and SSE | feature:live-docs-dashboard | depends:S-004,S-005 | plan:docs/plans/2026-03-07-live-docs-dashboard.md | spec:docs/features/2026-03-07-live-docs-dashboard.md

## Inbox
<!-- New items, not yet refined -->
