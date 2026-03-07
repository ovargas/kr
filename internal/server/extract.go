package server

import (
	"regexp"
	"strings"
)

var linkRe = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)

// extractMeta reads raw markdown bytes and returns the H1 title and a
// plain-text excerpt from the first paragraph after the title.
// Returns ("", "") if no H1 heading is found.
func extractMeta(data []byte) (title string, excerpt string) {
	lines := strings.Split(string(data), "\n")

	i := 0

	// Skip front matter
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		i++
		for i < len(lines) {
			if strings.TrimSpace(lines[i]) == "---" {
				i++
				break
			}
			i++
		}
	}

	// Find the first H1
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimSpace(line[2:])
			i++
			break
		}
		i++
	}

	if title == "" {
		return "", ""
	}

	// Collect excerpt lines: non-empty, non-heading lines after the title
	var excerptLines []string
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			if len(excerptLines) > 0 {
				break // end of first paragraph
			}
			i++
			continue
		}
		if strings.HasPrefix(line, "#") {
			break // next heading
		}
		excerptLines = append(excerptLines, line)
		i++
	}

	if len(excerptLines) == 0 {
		return title, ""
	}

	raw := strings.Join(excerptLines, " ")
	excerpt = stripMarkdown(raw)
	excerpt = truncate(excerpt, 200)

	return title, excerpt
}

// stripMarkdown removes basic inline markdown syntax from text.
func stripMarkdown(s string) string {
	// Remove blockquote prefix
	s = strings.TrimPrefix(s, "> ")

	// Replace [text](url) with text
	s = linkRe.ReplaceAllString(s, "$1")

	// Remove bold **text** and __text__
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")

	// Remove inline code backticks
	s = strings.ReplaceAll(s, "`", "")

	// Remove italic *text* — replace single * not adjacent to another *
	// Simple approach: already removed ** above, so remaining * are italic markers
	s = strings.ReplaceAll(s, "*", "")

	return strings.TrimSpace(s)
}

// truncate shortens s to maxLen characters at a word boundary, appending "..."
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Find the last space before maxLen
	cut := strings.LastIndex(s[:maxLen], " ")
	if cut <= 0 {
		cut = maxLen
	}
	return s[:cut] + "..."
}
