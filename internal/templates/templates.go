package templates

import (
	"embed"
	"html/template"
	"io"
)

//go:embed *.html
var templateFS embed.FS

// NavItem represents a folder link in the navigation bar.
type NavItem struct {
	Name   string
	Active bool
}

// Section represents a column in the Kanban board.
type Section struct {
	Name  string
	Items []BacklogItem
}

// BacklogItem represents a single card on the Kanban board.
type BacklogItem struct {
	Done   bool
	ID     string
	Title  string
	Fields []Field
	Status string
}

// Field represents a key-value pair on a backlog card.
type Field struct {
	Key      string
	Value    string
	IsLink   bool
	LinkPath string
}

// BacklogData is the template data for the Kanban board page.
type BacklogData struct {
	Sections []Section
	NavItems []NavItem
}

// FileEntry represents a file in a folder listing.
type FileEntry struct {
	Name    string
	Title   string
	Excerpt string
}

// FolderData is the template data for the folder listing page.
type FolderData struct {
	FolderName string
	Files      []FileEntry
	NavItems   []NavItem
}

// DocumentData is the template data for the document view page.
type DocumentData struct {
	FrontMatter map[string]any
	Content     template.HTML
	NavItems    []NavItem
}

// Templates holds parsed HTML templates and renders pages.
type Templates struct {
	layout   *template.Template
	backlog  *template.Template
	folder   *template.Template
	document *template.Template
}

// New parses all embedded templates and returns a Templates instance.
func New() (*Templates, error) {
	layout, err := template.ParseFS(templateFS, "layout.html")
	if err != nil {
		return nil, err
	}

	backlog, err := template.Must(layout.Clone()).ParseFS(templateFS, "backlog.html")
	if err != nil {
		return nil, err
	}

	folder, err := template.Must(layout.Clone()).ParseFS(templateFS, "folder.html")
	if err != nil {
		return nil, err
	}

	document, err := template.Must(layout.Clone()).ParseFS(templateFS, "document.html")
	if err != nil {
		return nil, err
	}

	return &Templates{
		layout:   layout,
		backlog:  backlog,
		folder:   folder,
		document: document,
	}, nil
}

// RenderBacklog renders the Kanban board page.
func (t *Templates) RenderBacklog(w io.Writer, data BacklogData) error {
	return t.backlog.ExecuteTemplate(w, "layout", data)
}

// RenderFolder renders the folder file listing page.
func (t *Templates) RenderFolder(w io.Writer, data FolderData) error {
	return t.folder.ExecuteTemplate(w, "layout", data)
}

// RenderDocument renders the markdown document view page.
func (t *Templates) RenderDocument(w io.Writer, data DocumentData) error {
	return t.document.ExecuteTemplate(w, "layout", data)
}
