package server

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ovargas/kr/internal/templates"
)

var knownFolders = []string{"bugs", "features", "decisions", "plans", "research", "reviews", "handoffs"}

// scanFolders returns nav items for all subdirectories under rootPath.
// Known folders appear first in fixed order, then unknown folders alphabetically.
func scanFolders(rootPath string) ([]templates.NavItem, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	known := make([]templates.NavItem, 0)
	unknown := make([]templates.NavItem, 0)

	dirSet := make(map[string]bool)
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		dirSet[e.Name()] = true
	}

	for _, name := range knownFolders {
		if dirSet[name] {
			known = append(known, templates.NavItem{Name: name})
			delete(dirSet, name)
		}
	}

	for name := range dirSet {
		unknown = append(unknown, templates.NavItem{Name: name})
	}
	slices.SortFunc(unknown, func(a, b templates.NavItem) int {
		return strings.Compare(a.Name, b.Name)
	})

	return append(known, unknown...), nil
}

// listFiles returns .md files in the given directory, sorted by name descending.
func listFiles(folderPath string) ([]templates.FileEntry, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var files []templates.FileEntry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		entry := templates.FileEntry{Name: e.Name()}
		if data, readErr := os.ReadFile(filepath.Join(folderPath, e.Name())); readErr == nil {
			entry.Title, entry.Excerpt = extractMeta(data)
		}
		files = append(files, entry)
	}

	slices.SortFunc(files, func(a, b templates.FileEntry) int {
		return strings.Compare(b.Name, a.Name) // descending
	})

	return files, nil
}

// scanFoldersWithActive returns nav items with the active folder marked.
func scanFoldersWithActive(rootPath, activeFolder string) ([]templates.NavItem, error) {
	items, err := scanFolders(rootPath)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Name == activeFolder {
			items[i].Active = true
		}
	}
	return items, nil
}

// isSubPath checks that resolved is inside base (prevents path traversal).
func isSubPath(base, resolved string) bool {
	rel, err := filepath.Rel(base, resolved)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..") && rel != ".."
}
