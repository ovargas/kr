package renderer

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

// Result holds the rendered HTML and extracted front matter from a markdown file.
type Result struct {
	HTML        string
	FrontMatter map[string]any
}

// Renderer converts markdown to HTML with front matter extraction.
type Renderer struct {
	md goldmark.Markdown
}

// New creates a Renderer with goldmark configured for GFM, front matter,
// auto heading IDs, and unsafe HTML passthrough.
func New() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&frontmatter.Extender{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	return &Renderer{md: md}
}

// Render converts markdown source bytes to HTML and extracts front matter.
func (r *Renderer) Render(source []byte) (*Result, error) {
	ctx := parser.NewContext()
	var buf bytes.Buffer

	if err := r.md.Convert(source, &buf, parser.WithContext(ctx)); err != nil {
		return nil, err
	}

	var meta map[string]any
	d := frontmatter.Get(ctx)
	if d != nil {
		if err := d.Decode(&meta); err != nil {
			return nil, err
		}
	}

	return &Result{
		HTML:        buf.String(),
		FrontMatter: meta,
	}, nil
}
