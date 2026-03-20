# kr

A Go CLI tool that launches a local web server to render and navigate Markdown documentation files. It provides a Kanban board view for `backlog.md`, folder-based navigation, and live auto-refresh when files change on disk.

## Features

- **Kanban board** — Renders `backlog.md` as a four-column board (Inbox, Ready, Doing, Done) with clickable links to referenced specs and plans
- **Folder navigation** — Top menu built from subdirectories, with known folders (features, plans, decisions, bugs, research, reviews, handoffs) listed first
- **Markdown rendering** — Converts `.md` files to styled HTML with GFM tables, auto-linked headings, and front matter display
- **Live reload** — Watches the documentation directory for changes and auto-refreshes the browser via Server-Sent Events (SSE)
- **Single binary** — All templates, CSS, and JS are embedded at compile time

## Installation

Requires **Go 1.25.0** or later.

```bash
# Install directly
go install github.com/ovargas/kr/cmd/kr@latest

# Or clone and build
git clone https://github.com/ovargas/kr.git
cd kr
go build -o kr ./cmd/kr
```

## Usage

```bash
# Serve the current directory on a random port
kr

# Serve a specific directory
kr --path /path/to/docs

# Use a specific port
kr --port 8080

# Both flags
kr --port 3000 --path ./my-project/docs
```

The server prints the URL to stdout:

```
listening on http://localhost:8080
```

Press `Ctrl+C` to stop the server.

### Flags

| Flag | Default | Description |
|---|---|---|
| `--port` | `0` (random) | Port to listen on |
| `--path` | `.` (current dir) | Path to documentation directory |

## Expected Directory Structure

Point `--path` at a directory containing Markdown files and folders:

```
docs/
├── backlog.md          # Rendered as a Kanban board at /
├── features/           # Shown in nav menu
│   ├── some-feature.md
│   └── another.md
├── plans/
│   └── implementation.md
├── decisions/
│   └── adr-001.md
├── bugs/
├── research/
├── reviews/
└── handoffs/
```

The known folders (`bugs`, `features`, `decisions`, `plans`, `research`, `reviews`, `handoffs`) appear first in the navigation. Any additional folders in the directory are listed alphabetically after them.

### Backlog Format

The `backlog.md` file is parsed into Kanban columns based on `## Heading` sections:

```markdown
# Backlog

## Inbox
- [ ] S-001: New item idea

## Ready
- [ ] S-002: Refined and planned item | feature:my-feature | spec:docs/features/spec.md

## Doing
- [>] S-003: Currently in progress | feature:active-work

## Done
- [x] S-004: Completed item | spec:docs/features/done.md — merged to main
```

Pipe-delimited fields (`key:value`) are displayed on each card. Fields containing `docs/` paths become clickable links.

## Routes

| Path | View |
|---|---|
| `/` | Kanban board (backlog.md) |
| `/{folder}` | File listing for that folder |
| `/{folder}/{file}.md` | Rendered Markdown document |
| `/events` | SSE endpoint for live reload |

## CI / Release

Releases are created automatically by [semantic-release](https://semantic-release.gitbook.io/) when commits following the [Conventional Commits](https://www.conventionalcommits.org/) format are merged into `main`.

### Required repository secret

| Secret name | Required permissions |
|---|---|
| `SEMANTIC_RELEASE_TOKEN` | A Personal Access Token (PAT) or GitHub App token with **Contents: Read & write** (and optionally Issues/PRs write for release notes). |

> **Why is this needed?**  Tags pushed by a workflow that uses the default `GITHUB_TOKEN` do not trigger other workflows (e.g. `release.yaml`). Using a PAT or GitHub App token makes the tag push behave like a normal user event and causes `release.yaml` to run and upload build artifacts to the GitHub Release.
>
> Without `SEMANTIC_RELEASE_TOKEN`, semantic-release will still create a tag and a GitHub Release, but the binary assets will **not** be attached.

### Manually re-running the Release workflow

If you need to re-attach assets to an existing tag, navigate to **Actions → Release → Run workflow** in the GitHub UI and select the appropriate tag. The `workflow_dispatch` trigger on `release.yaml` makes this possible.

## Development

```bash
# Run tests
go test ./...

# Build and run
go run ./cmd/kr --path ./docs

# Lint (requires golangci-lint)
golangci-lint run ./...
```

## CI / Release configuration

This repository uses two GitHub Actions workflows:

| Workflow | Trigger | Purpose |
|---|---|---|
| `semantic-release.yaml` | Push to `main` | Analyzes commits, bumps the version, and creates a Git tag + GitHub Release |
| `release.yaml` | Push of `v*` tags / manual dispatch | Builds multi-platform binaries and attaches them to the GitHub Release |

### Required repository secret — `GH_TOKEN`

The `semantic-release` workflow must authenticate with a **Personal Access Token (PAT)** or a **GitHub App token** stored as a repository secret named `GH_TOKEN`.

The default `GITHUB_TOKEN` provided to every workflow cannot trigger other workflows. When semantic-release pushes the new tag using `GITHUB_TOKEN`, GitHub treats that push as an internal action and `release.yaml` never fires. Using a PAT/App token makes the tag push look like a regular user/app event, which does trigger `release.yaml`.

**Required token scopes:**

| Scope (classic PAT) | Fine-grained permission | Why |
|---|---|---|
| `repo` | Contents: Read & write | Push tags and create/update releases |
| `write:discussion` (optional) | Discussions: Read & write (optional) | Required by some semantic-release plugins |

**How to add the secret:**

1. Create a [Personal Access Token](https://github.com/settings/tokens) (classic `repo` scope, or fine-grained with *Contents: Read & write* on this repository).
2. In the repository, go to **Settings → Secrets and variables → Actions → New repository secret**.
3. Name: `GH_TOKEN`, Value: your token.

After adding `GH_TOKEN`, the next merge to `main` will let semantic-release push a tag that automatically triggers the `Release` workflow to build and attach binaries.

### Manual release trigger

The `release.yaml` workflow also supports `workflow_dispatch`, so you can run it manually from the **Actions** tab for any existing tag (e.g. to re-attach binaries after an initial failed run).

## License

MIT
