# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**kr** is a Go CLI tool that launches a local web server to render and navigate Markdown documentation files in a formatted way.

- Module: `github.com/ovargas/kr`
- Go version: 1.25.0

## Build & Run Commands

```bash
# Build
go build -o kr .

# Run
go run . [flags]

# Run tests
go test ./...

# Run a single test
go test ./path/to/package -run TestName

# Lint (if golangci-lint is installed)
golangci-lint run ./...
```

## Architecture

The application is a single-binary CLI that:
1. Accepts a path to a directory containing `.md` files (or defaults to the current directory)
2. Starts an HTTP server on a local port
3. Renders Markdown files as styled HTML pages
4. Provides navigation between documents

### Key design decisions
- Standard library `net/http` for the web server (no heavy frameworks)
- Markdown-to-HTML rendering via a Go library (e.g., goldmark)
- Embedded templates/static assets using Go `embed` package
- CLI flag parsing with standard `flag` package or cobra
