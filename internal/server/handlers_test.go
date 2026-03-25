package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	dir := t.TempDir()

	// Create a folder with a .md file
	featDir := filepath.Join(dir, "features")
	if err := os.Mkdir(featDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mdContent := "---\ntitle: Test\n---\n\n# Hello\n\nWorld.\n"
	if err := os.WriteFile(filepath.Join(featDir, "test-doc.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatal(err)
	}

	srv, err := New(0, dir, "testproject")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv, dir
}

func TestRootURL(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	b, _ := io.ReadAll(resp.Body)
	body := string(b)
	if !strings.Contains(body, "kanban") {
		t.Error("response missing 'kanban' class")
	}
}

func TestFolderListing(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/features/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	b, _ := io.ReadAll(resp.Body)
	body := string(b)
	if !strings.Contains(body, "test-doc.md") {
		t.Error("response missing file name 'test-doc.md'")
	}
	if !strings.Contains(body, "Hello") {
		t.Error("response missing extracted title 'Hello'")
	}
	if !strings.Contains(body, "World.") {
		t.Error("response missing extracted excerpt 'World.'")
	}
}

func TestDocumentView(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/features/test-doc.md")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	b, _ := io.ReadAll(resp.Body)
	body := string(b)
	if !strings.Contains(body, "<h1") {
		t.Error("response missing rendered heading")
	}
	if !strings.Contains(body, "title") {
		t.Error("response missing front matter key")
	}
}

func TestMissingFolder(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestMissingFile(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/features/nonexistent.md")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPathTraversal(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/..%2F..%2Fetc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		t.Error("path traversal returned 200, expected non-200")
	}
}

func TestMissingBacklog(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	// No backlog.md in temp dir — should still return 200
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSearchBasic(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/search?q=World")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	body := string(b)
	if !strings.Contains(body, "Hello") {
		t.Error("response should contain title 'Hello' from matching file")
	}
	if !strings.Contains(body, "test-doc.md") {
		t.Error("response should contain filename 'test-doc.md'")
	}
}

func TestSearchNoResults(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/search?q=nonexistentterm12345")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	body := string(b)
	if !strings.Contains(body, "No files match") {
		t.Error("response should contain 'No files match' message")
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/search?q=")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	body := string(b)
	if !strings.Contains(body, "Enter a search term") {
		t.Error("response should contain empty state message")
	}
}

func TestSearchInvalidRegex(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/search?q=%5Binvalid&regex=1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200 (not 500)", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	body := string(b)
	if !strings.Contains(body, "Invalid regular expression") {
		t.Error("response should contain regex error message")
	}
}

func TestSearchCaseSensitive(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	// Case-insensitive (default): "world" should match "World."
	resp, err := http.Get(ts.URL + "/search?q=world")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(b), "test-doc.md") {
		t.Error("case-insensitive search for 'world' should find test-doc.md")
	}

	// Case-sensitive: "world" should NOT match "World."
	resp, err = http.Get(ts.URL + "/search?q=world&case=1")
	if err != nil {
		t.Fatal(err)
	}
	b, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if strings.Contains(string(b), "test-doc.md") {
		t.Error("case-sensitive search for 'world' should NOT find test-doc.md (file has 'World')")
	}
}

func TestSearchOverflow(t *testing.T) {
	dir := t.TempDir()
	folder := filepath.Join(dir, "docs")
	os.Mkdir(folder, 0o755)
	for i := 0; i < 55; i++ {
		name := filepath.Join(folder, fmt.Sprintf("file%03d.md", i))
		os.WriteFile(name, []byte("# Title\n\ncommonterm here\n"), 0o644)
	}

	srv, err := New(0, dir, "testproject")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/search?q=commonterm")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	body := string(b)
	if !strings.Contains(body, "Refine your search") {
		t.Error("response should contain overflow message")
	}
}
