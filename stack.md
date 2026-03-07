# Stack

## Project
- **Name:** kr
- **Description:** Go CLI tool that launches a local web server to render and navigate Markdown documentation files in a formatted way
- **Role:** cli
- **Hub:** standalone

## Language & Runtime
- **Language:** Go 1.25.0
- **Runtime:** single binary
- **Package manager:** Go modules

## Framework & Structure
- **Framework:** stdlib (`net/http`)
- **Structure:** layered: cmd + internal
- **Folder layout:**
  ```
  cmd/            # CLI entry point, flag parsing
  internal/
    server/       # HTTP server setup and routing
    renderer/     # Markdown-to-HTML rendering (goldmark)
    templates/    # Embedded HTML templates (Go embed)
    static/       # Embedded static assets — CSS, JS (Go embed)
  ```

## Data & Storage
- **Database:** N/A — filesystem only
- **ORM/Queries:** N/A
- **Migrations:** N/A
- **Cache:** N/A

## API & Communication
- **API style:** HTML only — server-rendered pages
- **Auth:** N/A — local tool
- **External services:** none

## Configuration
- **Config approach:** CLI flags only (standard `flag` package)
- **Environments:** local only (CLI tool)
- **Secrets:** N/A

## Testing & Quality
- **Test framework:** Go testing (`go test ./...`)
- **Lint:** golangci-lint
- **Type checking:** N/A (Go is statically typed)
- **CI/CD:** TBD

## Build & Deploy
- **Build:** `go build -o kr .`
- **Run locally:** `go run . [flags]`
- **Deploy target:** single binary distribution
- **Container:** no

## Key Libraries
- **Markdown rendering:** goldmark
- **Template/asset embedding:** Go `embed` package

## Decisions Made
<!-- Links to local ADRs in docs/decisions/ — starts empty -->

## TBD Items
- **CI/CD** — set up when ready to automate testing and releases
