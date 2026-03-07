package backlog

import (
	"strings"
)

// Backlog holds the parsed backlog sections.
type Backlog struct {
	Sections []Section
}

// Section represents a column (Inbox, Ready, Doing, Done).
type Section struct {
	Name  string
	Items []Item
}

// Item represents a single backlog entry.
type Item struct {
	Done   bool
	ID     string
	Title  string
	Fields []Field
	Status string
}

// Field represents a key-value pair on a backlog item.
type Field struct {
	Key      string
	Value    string
	IsLink   bool
	LinkPath string
}

var sectionOrder = []string{"Inbox", "Ready", "Doing", "Done"}

// Parse converts backlog markdown content into structured data.
func Parse(content []byte) (*Backlog, error) {
	lines := strings.Split(string(content), "\n")

	sectionMap := make(map[string][]Item)
	var currentSection string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			currentSection = strings.TrimPrefix(trimmed, "## ")
			continue
		}

		if currentSection == "" {
			continue
		}

		if !strings.HasPrefix(trimmed, "- [") {
			continue
		}

		item := parseLine(trimmed)
		sectionMap[currentSection] = append(sectionMap[currentSection], item)
	}

	var sections []Section
	for _, name := range sectionOrder {
		sections = append(sections, Section{
			Name:  name,
			Items: sectionMap[name],
		})
	}

	return &Backlog{Sections: sections}, nil
}

func parseLine(line string) Item {
	var item Item

	// Extract checkbox state
	if strings.HasPrefix(line, "- [x]") || strings.HasPrefix(line, "- [X]") {
		item.Done = true
	}

	// Remove prefix "- [x] " or "- [ ] " or "- [>] " or "- [=] "
	text := line
	if idx := strings.Index(text, "] "); idx >= 0 {
		text = text[idx+2:]
	}

	// Split on " — " (em dash) to extract status note
	if parts := strings.SplitN(text, " — ", 2); len(parts) == 2 {
		text = parts[0]
		item.Status = strings.TrimSpace(parts[1])
	}

	// Split on " | " to separate ID:Title from fields
	segments := strings.Split(text, " | ")

	// First segment is "ID: Title"
	if len(segments) > 0 {
		idTitle := strings.TrimSpace(segments[0])
		if colonIdx := strings.Index(idTitle, ": "); colonIdx >= 0 {
			item.ID = idTitle[:colonIdx]
			item.Title = idTitle[colonIdx+2:]
		} else {
			item.Title = idTitle
		}
	}

	// Remaining segments are fields
	for _, seg := range segments[1:] {
		seg = strings.TrimSpace(seg)
		field := parseField(seg)
		item.Fields = append(item.Fields, field)
	}

	return item
}

func parseField(s string) Field {
	var f Field
	if idx := strings.Index(s, ":"); idx >= 0 {
		f.Key = strings.TrimSpace(s[:idx])
		f.Value = strings.TrimSpace(s[idx+1:])
	} else {
		f.Value = s
	}

	if strings.Contains(f.Value, "docs/") {
		f.IsLink = true
		// Strip "docs/" prefix to make internal viewer path
		if idx := strings.Index(f.Value, "docs/"); idx >= 0 {
			f.LinkPath = "/" + f.Value[idx+5:]
		}
	}

	return f
}
