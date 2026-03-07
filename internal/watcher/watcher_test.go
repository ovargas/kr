package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileModification(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.md")
	os.WriteFile(file, []byte("hello"), 0o644)

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	os.WriteFile(file, []byte("world"), 0o644)

	select {
	case <-w.Changes():
		// OK
	case <-time.After(500 * time.Millisecond):
		t.Error("did not receive change event after file modification")
	}
}

func TestFileCreation(t *testing.T) {
	dir := t.TempDir()

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	time.Sleep(50 * time.Millisecond)

	os.WriteFile(filepath.Join(dir, "new.md"), []byte("content"), 0o644)

	select {
	case <-w.Changes():
		// OK
	case <-time.After(500 * time.Millisecond):
		t.Error("did not receive change event after file creation")
	}
}

func TestDebouncing(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.md")
	os.WriteFile(file, []byte("hello"), 0o644)

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	time.Sleep(50 * time.Millisecond)

	// Rapidly modify 5 times
	for i := 0; i < 5; i++ {
		os.WriteFile(file, []byte("change"), 0o644)
		time.Sleep(10 * time.Millisecond)
	}

	// Should receive exactly 1 event
	select {
	case <-w.Changes():
		// First event received
	case <-time.After(500 * time.Millisecond):
		t.Error("did not receive debounced event")
	}

	// Should NOT receive a second event
	select {
	case <-w.Changes():
		t.Error("received extra event — debouncing failed")
	case <-time.After(300 * time.Millisecond):
		// OK — no second event
	}
}

func TestNonMdFilesIgnored(t *testing.T) {
	dir := t.TempDir()

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	time.Sleep(50 * time.Millisecond)

	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0o644)

	select {
	case <-w.Changes():
		t.Error("received event for non-.md file")
	case <-time.After(300 * time.Millisecond):
		// OK — no event for .txt file
	}
}

func TestSubdirectoryWatching(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "features")
	os.Mkdir(sub, 0o755)

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	time.Sleep(50 * time.Millisecond)

	os.WriteFile(filepath.Join(sub, "doc.md"), []byte("content"), 0o644)

	select {
	case <-w.Changes():
		// OK
	case <-time.After(500 * time.Millisecond):
		t.Error("did not receive event for subdirectory file")
	}
}
