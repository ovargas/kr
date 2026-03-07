package backlog

import (
	"testing"
)

const sampleBacklog = `# Backlog

## Doing
- [>] S-003: Story three | feature:fix | service:be

## Ready
- [ ] S-004: Story four | feature:nav | depends:S-002,S-003
- [ ] S-005: Story five | feature:nav | depends:S-003

## Done
- [x] S-001: Story one | feature:fix | plan:docs/plans/file.md | spec:docs/features/spec.md
- [x] S-002: Story two | feature:render

## Inbox
- [ ] S-006: Story six
`

func TestFullBacklogParsing(t *testing.T) {
	b, err := Parse([]byte(sampleBacklog))
	if err != nil {
		t.Fatal(err)
	}

	if len(b.Sections) != 4 {
		t.Fatalf("sections = %d, want 4", len(b.Sections))
	}

	// Check section ordering: Inbox, Ready, Doing, Done
	names := []string{"Inbox", "Ready", "Doing", "Done"}
	for i, want := range names {
		if b.Sections[i].Name != want {
			t.Errorf("section[%d] = %q, want %q", i, b.Sections[i].Name, want)
		}
	}

	// Check item counts
	counts := map[string]int{"Inbox": 1, "Ready": 2, "Doing": 1, "Done": 2}
	for _, s := range b.Sections {
		if len(s.Items) != counts[s.Name] {
			t.Errorf("%s items = %d, want %d", s.Name, len(s.Items), counts[s.Name])
		}
	}
}

func TestItemFieldExtraction(t *testing.T) {
	b, _ := Parse([]byte(sampleBacklog))

	// S-001 in Done section
	done := b.Sections[3] // Done
	if len(done.Items) < 1 {
		t.Fatal("no items in Done")
	}
	item := done.Items[0]

	if item.ID != "S-001" {
		t.Errorf("ID = %q, want S-001", item.ID)
	}
	if item.Title != "Story one" {
		t.Errorf("Title = %q, want 'Story one'", item.Title)
	}
	if len(item.Fields) != 3 {
		t.Errorf("fields = %d, want 3", len(item.Fields))
	}
}

func TestCheckboxState(t *testing.T) {
	b, _ := Parse([]byte(sampleBacklog))

	done := b.Sections[3] // Done
	if !done.Items[0].Done {
		t.Error("S-001 should be done")
	}

	ready := b.Sections[1] // Ready
	if ready.Items[0].Done {
		t.Error("S-004 should not be done")
	}
}

func TestStatusNote(t *testing.T) {
	input := `## Doing
- [>] S-003: Story | feature:x — in progress on main
`
	b, _ := Parse([]byte(input))
	doing := b.Sections[2] // Doing
	if len(doing.Items) == 0 {
		t.Fatal("no items in Doing")
	}
	if doing.Items[0].Status != "in progress on main" {
		t.Errorf("Status = %q, want 'in progress on main'", doing.Items[0].Status)
	}
}

func TestMissingSections(t *testing.T) {
	input := `## Ready
- [ ] S-001: Story one

## Done
- [x] S-002: Story two
`
	b, _ := Parse([]byte(input))
	if len(b.Sections) != 4 {
		t.Fatalf("sections = %d, want 4", len(b.Sections))
	}
	// Inbox and Doing should be empty
	if len(b.Sections[0].Items) != 0 {
		t.Error("Inbox should be empty")
	}
	if len(b.Sections[2].Items) != 0 {
		t.Error("Doing should be empty")
	}
}

func TestEmptyBacklog(t *testing.T) {
	input := `# Backlog

## Doing

## Ready

## Done

## Inbox
`
	b, _ := Parse([]byte(input))
	for _, s := range b.Sections {
		if len(s.Items) != 0 {
			t.Errorf("%s should be empty, has %d items", s.Name, len(s.Items))
		}
	}
}

func TestLinkDetection(t *testing.T) {
	tests := []struct {
		input    string
		isLink   bool
		linkPath string
	}{
		{"plan:docs/plans/2026-03-06-file.md", true, "/plans/2026-03-06-file.md"},
		{"spec:docs/features/2026-03-06-file.md", true, "/features/2026-03-06-file.md"},
		{"epic:docs/epics/2026-03-07-epic.md", true, "/epics/2026-03-07-epic.md"},
		{"service:be", false, ""},
		{"feature:nav", false, ""},
	}

	for _, tc := range tests {
		f := parseField(tc.input)
		if f.IsLink != tc.isLink {
			t.Errorf("%q: IsLink = %v, want %v", tc.input, f.IsLink, tc.isLink)
		}
		if f.LinkPath != tc.linkPath {
			t.Errorf("%q: LinkPath = %q, want %q", tc.input, f.LinkPath, tc.linkPath)
		}
	}
}
