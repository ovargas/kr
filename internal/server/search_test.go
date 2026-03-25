package server

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCompilePattern_CaseInsensitive(t *testing.T) {
	re, err := compilePattern(SearchOptions{Query: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if !re.MatchString("Hello World") {
		t.Error("case-insensitive pattern should match 'Hello World'")
	}
}

func TestCompilePattern_CaseSensitive(t *testing.T) {
	re, err := compilePattern(SearchOptions{Query: "hello", CaseSensitive: true})
	if err != nil {
		t.Fatal(err)
	}
	if re.MatchString("Hello World") {
		t.Error("case-sensitive 'hello' should NOT match 'Hello World'")
	}
	if !re.MatchString("hello world") {
		t.Error("case-sensitive 'hello' should match 'hello world'")
	}
}

func TestCompilePattern_WholeWord(t *testing.T) {
	re, err := compilePattern(SearchOptions{Query: "plan", WholeWord: true})
	if err != nil {
		t.Fatal(err)
	}
	if !re.MatchString("the plan works") {
		t.Error("whole word 'plan' should match 'the plan works'")
	}
	if re.MatchString("the plans work") {
		t.Error("whole word 'plan' should NOT match 'the plans work'")
	}
}

func TestCompilePattern_Regex(t *testing.T) {
	re, err := compilePattern(SearchOptions{Query: "feat.*", UseRegex: true})
	if err != nil {
		t.Fatal(err)
	}
	if !re.MatchString("this is a feature") {
		t.Error("regex 'feat.*' should match 'this is a feature'")
	}
}

func TestCompilePattern_InvalidRegex(t *testing.T) {
	_, err := compilePattern(SearchOptions{Query: "[invalid", UseRegex: true})
	if err == nil {
		t.Error("invalid regex should return error")
	}
}

func TestCompilePattern_SpecialCharsEscaped(t *testing.T) {
	re, err := compilePattern(SearchOptions{Query: "foo.bar"})
	if err != nil {
		t.Fatal(err)
	}
	if re.MatchString("fooXbar") {
		t.Error("plain text 'foo.bar' should NOT match 'fooXbar' (dot should be escaped)")
	}
	if !re.MatchString("foo.bar") {
		t.Error("plain text 'foo.bar' should match literal 'foo.bar'")
	}
}

func setupSearchDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// features/doc1.md — contains "authentication"
	featDir := filepath.Join(dir, "features")
	os.Mkdir(featDir, 0o755)
	os.WriteFile(filepath.Join(featDir, "doc1.md"), []byte("# Auth Feature\n\nThis covers authentication.\n"), 0o644)

	// features/doc2.md — contains "authorization"
	os.WriteFile(filepath.Join(featDir, "doc2.md"), []byte("# Auth Rules\n\nThis covers authorization.\n"), 0o644)

	// plans/plan1.md — contains "Bug fix plan"
	planDir := filepath.Join(dir, "plans")
	os.Mkdir(planDir, 0o755)
	os.WriteFile(filepath.Join(planDir, "plan1.md"), []byte("# Bug Fix\n\n## Problem\n\nA Bug was found.\n"), 0o644)

	// root-level file
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Readme\n\nProject overview.\n"), 0o644)

	return dir
}

func TestMatchesFile_Found(t *testing.T) {
	dir := setupSearchDir(t)
	re, _ := compilePattern(SearchOptions{Query: "authentication"})
	if !matchesFile(filepath.Join(dir, "features", "doc1.md"), re) {
		t.Error("should find 'authentication' in doc1.md")
	}
}

func TestMatchesFile_NotFound(t *testing.T) {
	dir := setupSearchDir(t)
	re, _ := compilePattern(SearchOptions{Query: "nonexistent"})
	if matchesFile(filepath.Join(dir, "features", "doc1.md"), re) {
		t.Error("should NOT find 'nonexistent' in doc1.md")
	}
}

func TestSearchFiles_BasicMatch(t *testing.T) {
	dir := setupSearchDir(t)
	outcome, err := searchFiles(dir, SearchOptions{Query: "authentication"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(outcome.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(outcome.Results))
	}
	r := outcome.Results[0]
	if r.Folder != "features" {
		t.Errorf("folder = %q, want 'features'", r.Folder)
	}
	if r.Name != "doc1.md" {
		t.Errorf("name = %q, want 'doc1.md'", r.Name)
	}
	if r.Title != "Auth Feature" {
		t.Errorf("title = %q, want 'Auth Feature'", r.Title)
	}
}

func TestSearchFiles_MultipleResults(t *testing.T) {
	dir := setupSearchDir(t)
	outcome, err := searchFiles(dir, SearchOptions{Query: "auth"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(outcome.Results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(outcome.Results))
	}
}

func TestSearchFiles_EmptyQuery(t *testing.T) {
	dir := setupSearchDir(t)
	outcome, err := searchFiles(dir, SearchOptions{Query: ""}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(outcome.Results) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(outcome.Results))
	}
}

func TestSearchFiles_CaseSensitive(t *testing.T) {
	dir := setupSearchDir(t)

	// Case-insensitive: "bug" should match "Bug"
	outcome, err := searchFiles(dir, SearchOptions{Query: "bug"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	insensitiveCount := len(outcome.Results)

	// Case-sensitive: "bug" should NOT match "Bug"
	outcome, err = searchFiles(dir, SearchOptions{Query: "bug", CaseSensitive: true}, 50)
	if err != nil {
		t.Fatal(err)
	}
	sensitiveCount := len(outcome.Results)

	if sensitiveCount >= insensitiveCount {
		t.Errorf("case-sensitive (%d) should find fewer results than case-insensitive (%d)", sensitiveCount, insensitiveCount)
	}
}

func TestSearchFiles_WholeWord(t *testing.T) {
	dir := setupSearchDir(t)

	// Without whole word: "auth" matches "authentication" and "authorization"
	outcome, err := searchFiles(dir, SearchOptions{Query: "auth"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	partialCount := len(outcome.Results)

	// With whole word: "auth" should NOT match "authentication"
	outcome, err = searchFiles(dir, SearchOptions{Query: "authentication", WholeWord: true}, 50)
	if err != nil {
		t.Fatal(err)
	}
	wholeCount := len(outcome.Results)

	if partialCount <= wholeCount {
		t.Errorf("partial match (%d) should find more than whole-word match (%d)", partialCount, wholeCount)
	}
}

func TestSearchFiles_Regex(t *testing.T) {
	dir := setupSearchDir(t)
	outcome, err := searchFiles(dir, SearchOptions{Query: "^## Problem", UseRegex: true}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(outcome.Results) != 1 {
		t.Fatalf("expected 1 result for regex '^## Problem', got %d", len(outcome.Results))
	}
	if outcome.Results[0].Name != "plan1.md" {
		t.Errorf("expected plan1.md, got %s", outcome.Results[0].Name)
	}
}

func TestSearchFiles_Overflow(t *testing.T) {
	dir := t.TempDir()
	folder := filepath.Join(dir, "docs")
	os.Mkdir(folder, 0o755)

	// Create 55 files all containing "match"
	for i := 0; i < 55; i++ {
		name := filepath.Join(folder, fmt.Sprintf("file%03d.md", i))
		os.WriteFile(name, []byte("# Title\n\nmatch here\n"), 0o644)
	}

	outcome, err := searchFiles(dir, SearchOptions{Query: "match"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(outcome.Results) != 50 {
		t.Errorf("expected 50 results, got %d", len(outcome.Results))
	}
	if !outcome.Overflow {
		t.Error("expected overflow to be true")
	}
}

func TestSearchFiles_InvalidRegex(t *testing.T) {
	dir := setupSearchDir(t)
	_, err := searchFiles(dir, SearchOptions{Query: "[invalid", UseRegex: true}, 50)
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestSearchFiles_RootLevelFile(t *testing.T) {
	dir := setupSearchDir(t)
	outcome, err := searchFiles(dir, SearchOptions{Query: "overview"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range outcome.Results {
		if r.Name == "readme.md" && r.Folder == "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should find readme.md at root level with empty folder")
	}
}
