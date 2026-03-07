package renderer

import (
	"strings"
	"testing"
)

func TestBasicMarkdown(t *testing.T) {
	r := New()
	src := []byte("# Hello\n\nA paragraph with **bold** and *italic*.\n\n- item 1\n- item 2\n")

	result, err := r.Render(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{"<h1", "Hello", "<strong>bold</strong>", "<em>italic</em>", "<li>item 1</li>"}
	for _, want := range checks {
		if !strings.Contains(result.HTML, want) {
			t.Errorf("HTML missing %q:\n%s", want, result.HTML)
		}
	}
}

func TestFrontMatterExtraction(t *testing.T) {
	r := New()
	src := []byte("---\ntitle: My Doc\nstatus: draft\ntags: [a, b]\n---\n\n# Content\n")

	result, err := r.Render(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FrontMatter == nil {
		t.Fatal("expected front matter, got nil")
	}
	if result.FrontMatter["title"] != "My Doc" {
		t.Errorf("title = %v, want %q", result.FrontMatter["title"], "My Doc")
	}
	if result.FrontMatter["status"] != "draft" {
		t.Errorf("status = %v, want %q", result.FrontMatter["status"], "draft")
	}
	if !strings.Contains(result.HTML, "<h1") {
		t.Errorf("HTML missing heading:\n%s", result.HTML)
	}
}

func TestNoFrontMatter(t *testing.T) {
	r := New()
	src := []byte("# Just a heading\n\nSome text.\n")

	result, err := r.Render(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.FrontMatter) != 0 {
		t.Errorf("expected no front matter, got %v", result.FrontMatter)
	}
	if !strings.Contains(result.HTML, "<h1") {
		t.Errorf("HTML missing heading:\n%s", result.HTML)
	}
}

func TestGFMTables(t *testing.T) {
	r := New()
	src := []byte("| A | B |\n|---|---|\n| 1 | 2 |\n")

	result, err := r.Render(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.HTML, "<table>") {
		t.Errorf("HTML missing <table>:\n%s", result.HTML)
	}
}

func TestEmptyInput(t *testing.T) {
	r := New()
	result, err := r.Render([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.HTML != "" {
		t.Errorf("expected empty HTML, got %q", result.HTML)
	}
}
