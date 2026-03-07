---
id: BUG-001
date: 2026-03-07
severity: medium
status: fixed
feature: FEAT-002
reported_by: founder
---

# Bug: Excerpt Shows Content Before First Section, Not First Section's Content

## Summary

The file listing excerpt is drawn from the content between the `# Title` and the first `##` heading — typically the `> **Value:** ...` blockquote in structured docs — rather than the content body of the first named section.

## Observed Behavior

For a document structured like this:

```
# Feature Name

> **Value:** Makes folder browsing actually useful...

## Problem

AI agents continuously update project documentation, but there's no way to monitor...
```

The excerpt shown in the folder listing is:

> **Value:** Makes folder browsing actually useful...

(The blockquote between the H1 and the first `##` heading.)

## Expected Behavior

The excerpt should be drawn from the content of the first `##` section (or whichever `##`-level heading comes first after the title). For the example above, the expected excerpt is:

> AI agents continuously update project documentation, but there's no way to monitor...

This is the meaningful description of what the document is about.

## Reproduction Steps

1. Open `kr` and navigate to a folder containing structured markdown files (e.g., `features/`)
2. Observe the excerpt shown under each file's title
3. Click through to the document and compare the excerpt to the actual first section content

**Reproduction rate:** Always — any document following the standard structure (H1 → blockquote value → `##` sections) exhibits this.

## Environment

- All structured markdown files following the project's standard format
- Feature specs, plans, bugs, decisions — any document with a `> **Value:**` line before the first `##` heading

## Impact

- **Who's affected:** All users browsing any folder with structured docs
- **Workaround:** Click into each file to read the actual content — defeats the purpose of the excerpt
- **Business impact:** The excerpt shows a generic tagline/value statement rather than the specific description of the document's content. When browsing many similar files, all excerpts start with "Makes..." which makes them indistinguishable.

## Initial Hypothesis

The issue is in `internal/server/extract.go` in the `extractMeta()` function. After finding the H1 title, it collects lines until the first blank line gap or `##` heading — which captures the blockquote `> **Value:** ...` line that sits between the title and the first `##` section.

The fix should skip over content before the first `##` heading and instead collect content from *within* that first `##` section.

## Related

- Feature spec: `docs/features/2026-03-07-rich-file-listing.md` (FEAT-002)
- Implementation: `internal/server/extract.go` — `extractMeta()` function
- Tests: `internal/server/extract_test.go` — will need to add/update test cases
