package server

import (
	"strings"
	"testing"
)

func TestExtractMeta_StandardFile(t *testing.T) {
	data := []byte("---\ntitle: Test\nstatus: draft\n---\n\n# My Document\n\nThis is the first paragraph of content.\n")
	title, excerpt := extractMeta(data)

	if title != "My Document" {
		t.Errorf("title = %q, want %q", title, "My Document")
	}
	if excerpt != "This is the first paragraph of content." {
		t.Errorf("excerpt = %q, want %q", excerpt, "This is the first paragraph of content.")
	}
}

func TestExtractMeta_NoFrontMatter(t *testing.T) {
	data := []byte("# Hello World\n\nSome text here.\n")
	title, excerpt := extractMeta(data)

	if title != "Hello World" {
		t.Errorf("title = %q, want %q", title, "Hello World")
	}
	if excerpt != "Some text here." {
		t.Errorf("excerpt = %q, want %q", excerpt, "Some text here.")
	}
}

func TestExtractMeta_NoH1(t *testing.T) {
	data := []byte("## Not an H1\n\nSome content.\n")
	title, excerpt := extractMeta(data)

	if title != "" {
		t.Errorf("title = %q, want empty", title)
	}
	if excerpt != "" {
		t.Errorf("excerpt = %q, want empty", excerpt)
	}
}

func TestExtractMeta_EmptyFile(t *testing.T) {
	title, excerpt := extractMeta([]byte{})

	if title != "" {
		t.Errorf("title = %q, want empty", title)
	}
	if excerpt != "" {
		t.Errorf("excerpt = %q, want empty", excerpt)
	}
}

func TestExtractMeta_H1NoContent(t *testing.T) {
	data := []byte("# Title Only\n")
	title, excerpt := extractMeta(data)

	if title != "Title Only" {
		t.Errorf("title = %q, want %q", title, "Title Only")
	}
	if excerpt != "" {
		t.Errorf("excerpt = %q, want empty", excerpt)
	}
}

func TestExtractMeta_MarkdownStripping(t *testing.T) {
	data := []byte("# Doc\n\nThis has **bold** and *italic* and `code` and [a link](http://example.com) in it.\n")
	title, excerpt := extractMeta(data)

	if title != "Doc" {
		t.Errorf("title = %q, want %q", title, "Doc")
	}

	want := "This has bold and italic and code and a link in it."
	if excerpt != want {
		t.Errorf("excerpt = %q, want %q", excerpt, want)
	}
}

func TestExtractMeta_LongExcerpt(t *testing.T) {
	long := strings.Repeat("word ", 50) // 250 chars
	data := []byte("# Title\n\n" + long + "\n")
	title, excerpt := extractMeta(data)

	if title != "Title" {
		t.Errorf("title = %q, want %q", title, "Title")
	}
	if len(excerpt) > 203 { // 200 + "..."
		t.Errorf("excerpt length = %d, want <= 203", len(excerpt))
	}
	if !strings.HasSuffix(excerpt, "...") {
		t.Errorf("excerpt should end with '...', got %q", excerpt[len(excerpt)-10:])
	}
}

func TestExtractMeta_FrontMatterOnly(t *testing.T) {
	data := []byte("---\ntitle: Test\n---\n")
	title, excerpt := extractMeta(data)

	if title != "" {
		t.Errorf("title = %q, want empty", title)
	}
	if excerpt != "" {
		t.Errorf("excerpt = %q, want empty", excerpt)
	}
}

func TestExtractMeta_BlockquoteValueWithSection(t *testing.T) {
	data := []byte("# Feature\n\n> **Value:** Some important text here.\n\n## Problem\n\nThe actual problem description lives here.\n")
	title, excerpt := extractMeta(data)

	if title != "Feature" {
		t.Errorf("title = %q, want %q", title, "Feature")
	}

	want := "The actual problem description lives here."
	if excerpt != want {
		t.Errorf("excerpt = %q, want %q", excerpt, want)
	}
}

func TestExtractMeta_BlockquoteValueNoSection(t *testing.T) {
	// No ## section — falls back to content after H1
	data := []byte("# Feature\n\n> **Value:** Some important text here.\n")
	title, excerpt := extractMeta(data)

	if title != "Feature" {
		t.Errorf("title = %q, want %q", title, "Feature")
	}

	want := "Value: Some important text here."
	if excerpt != want {
		t.Errorf("excerpt = %q, want %q", excerpt, want)
	}
}

func TestExtractMeta_H1FollowedByHeading(t *testing.T) {
	data := []byte("# Main Title\n\n## Section\n\nContent under section.\n")
	title, excerpt := extractMeta(data)

	if title != "Main Title" {
		t.Errorf("title = %q, want %q", title, "Main Title")
	}
	want := "Content under section."
	if excerpt != want {
		t.Errorf("excerpt = %q, want %q", excerpt, want)
	}
}

func TestExtractMeta_StructuredDoc(t *testing.T) {
	data := []byte("---\nid: FEAT-001\nstatus: draft\n---\n\n# Live Documentation Dashboard\n\n> **Value:** Gives you a real-time control panel.\n\n## Problem\n\nAI agents continuously update project documentation, but there's no way to monitor this activity in real time.\n")
	title, excerpt := extractMeta(data)

	if title != "Live Documentation Dashboard" {
		t.Errorf("title = %q, want %q", title, "Live Documentation Dashboard")
	}

	want := "AI agents continuously update project documentation, but there's no way to monitor this activity in real time."
	if excerpt != want {
		t.Errorf("excerpt = %q, want %q", excerpt, want)
	}
}
