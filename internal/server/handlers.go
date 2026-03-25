package server

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ovargas/kr/internal/backlog"
	"github.com/ovargas/kr/internal/templates"
)

func (s *Server) handleBacklog(w http.ResponseWriter, r *http.Request) {
	navItems, err := scanFolders(s.rootPath)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := templates.BacklogData{
		ProjectName: s.projectName,
		NavItems:    navItems,
	}

	// Read and parse backlog.md — missing file renders empty board
	content, err := os.ReadFile(filepath.Join(s.rootPath, "backlog.md"))
	if err == nil {
		b, parseErr := backlog.Parse(content)
		if parseErr == nil {
			for _, sec := range b.Sections {
				var items []templates.BacklogItem
				for _, it := range sec.Items {
					var fields []templates.Field
					for _, f := range it.Fields {
						fields = append(fields, templates.Field{
							Key:      f.Key,
							Value:    f.Value,
							IsLink:   f.IsLink,
							LinkPath: f.LinkPath,
						})
					}
					items = append(items, templates.BacklogItem{
						Done:   it.Done,
						ID:     it.ID,
						Title:  it.Title,
						Fields: fields,
						Status: it.Status,
					})
				}
				data.Sections = append(data.Sections, templates.Section{
					Name:  sec.Name,
					Items: items,
				})
			}
		}
	}

	if err := s.tmpl.RenderBacklog(w, data); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (s *Server) handleFolder(w http.ResponseWriter, r *http.Request) {
	folder := r.PathValue("folder")
	folderPath := filepath.Join(s.rootPath, folder)
	folderPath = filepath.Clean(folderPath)

	if !isSubPath(s.rootPath, folderPath) {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(folderPath)
	if err != nil || !info.IsDir() {
		http.NotFound(w, r)
		return
	}

	files, err := listFiles(folderPath)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	navItems, err := scanFoldersWithActive(s.rootPath, folder)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := templates.FolderData{
		ProjectName: s.projectName,
		FolderName:  folder,
		Files:       files,
		NavItems:    navItems,
	}

	if err := s.tmpl.RenderFolder(w, data); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (s *Server) handleDocument(w http.ResponseWriter, r *http.Request) {
	folder := r.PathValue("folder")
	file := r.PathValue("file")

	if !strings.HasSuffix(file, ".md") {
		http.NotFound(w, r)
		return
	}

	filePath := filepath.Join(s.rootPath, folder, file)
	filePath = filepath.Clean(filePath)

	if !isSubPath(s.rootPath, filePath) {
		http.NotFound(w, r)
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	result, err := s.renderer.Render(content)
	if err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}

	navItems, err := scanFoldersWithActive(s.rootPath, folder)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := templates.DocumentData{
		ProjectName: s.projectName,
		FrontMatter: result.FrontMatter,
		Content:     template.HTML(result.HTML),
		NavItems:    navItems,
	}

	if err := s.tmpl.RenderDocument(w, data); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

const maxSearchResults = 50

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	caseSensitive := r.URL.Query().Get("case") == "1"
	wholeWord := r.URL.Query().Get("word") == "1"
	useRegex := r.URL.Query().Get("regex") == "1"

	navItems, err := scanFolders(s.rootPath)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := templates.SearchData{
		ProjectName:   s.projectName,
		NavItems:      navItems,
		Query:         query,
		CaseSensitive: caseSensitive,
		WholeWord:     wholeWord,
		UseRegex:      useRegex,
		SearchQuery:   query,
		SearchCase:    caseSensitive,
		SearchWord:    wholeWord,
		SearchRegex:   useRegex,
	}

	if query != "" {
		opts := SearchOptions{
			Query:         query,
			CaseSensitive: caseSensitive,
			WholeWord:     wholeWord,
			UseRegex:      useRegex,
		}
		outcome, searchErr := searchFiles(s.rootPath, opts, maxSearchResults)
		if searchErr != nil {
			data.Error = "Invalid regular expression: " + searchErr.Error()
		} else {
			for _, r := range outcome.Results {
				data.Results = append(data.Results, templates.SearchResultEntry{
					Folder:  r.Folder,
					Name:    r.Name,
					Title:   r.Title,
					Excerpt: r.Excerpt,
				})
			}
			data.TotalFound = len(outcome.Results)
			if outcome.Overflow {
				data.TotalFound = maxSearchResults
			}
			data.Overflow = outcome.Overflow
		}
	}

	if err := s.tmpl.RenderSearch(w, data); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
