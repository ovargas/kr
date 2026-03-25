package server

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchOptions holds the parameters for a content search.
type SearchOptions struct {
	Query         string
	CaseSensitive bool
	WholeWord     bool
	UseRegex      bool
}

// SearchResult represents a single file that matched the search.
type SearchResult struct {
	Folder  string
	Name    string
	Title   string
	Excerpt string
}

// SearchOutcome wraps search results with an overflow indicator.
type SearchOutcome struct {
	Results  []SearchResult
	Overflow bool
}

// compilePattern builds a regexp from the search options.
func compilePattern(opts SearchOptions) (*regexp.Regexp, error) {
	pattern := opts.Query
	if !opts.UseRegex {
		pattern = regexp.QuoteMeta(pattern)
	}
	if opts.WholeWord {
		pattern = `\b` + pattern + `\b`
	}
	if !opts.CaseSensitive {
		pattern = "(?i)" + pattern
	}
	return regexp.Compile(pattern)
}

// matchesFile returns true if any line in the file matches the regexp.
func matchesFile(filePath string, re *regexp.Regexp) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if re.MatchString(scanner.Text()) {
			return true
		}
	}
	return false
}

// searchFiles walks rootPath recursively, searching .md files for matches.
// It stops after maxResults matches and sets Overflow if more exist.
func searchFiles(rootPath string, opts SearchOptions, maxResults int) (SearchOutcome, error) {
	if opts.Query == "" {
		return SearchOutcome{}, nil
	}

	re, err := compilePattern(opts)
	if err != nil {
		return SearchOutcome{}, err
	}

	var outcome SearchOutcome

	walkErr := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip entries with errors
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		if !matchesFile(path, re) {
			return nil
		}

		dir := filepath.Dir(path)
		folder, _ := filepath.Rel(rootPath, dir)
		if folder == "." {
			folder = ""
		}

		var title, excerpt string
		if data, readErr := os.ReadFile(path); readErr == nil {
			title, excerpt = extractMeta(data)
		}

		outcome.Results = append(outcome.Results, SearchResult{
			Folder:  folder,
			Name:    d.Name(),
			Title:   title,
			Excerpt: excerpt,
		})

		if len(outcome.Results) > maxResults {
			outcome.Overflow = true
			outcome.Results = outcome.Results[:maxResults]
			return filepath.SkipAll
		}

		return nil
	})

	if walkErr != nil {
		return SearchOutcome{}, walkErr
	}

	return outcome, nil
}
