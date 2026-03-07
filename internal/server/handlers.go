package server

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ovargas/kr/internal/templates"
)

func (s *Server) handleBacklog(w http.ResponseWriter, r *http.Request) {
	navItems, err := scanFolders(s.rootPath)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := templates.BacklogData{
		NavItems: navItems,
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
		FolderName: folder,
		Files:      files,
		NavItems:   navItems,
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
		FrontMatter: result.FrontMatter,
		Content:     template.HTML(result.HTML),
		NavItems:    navItems,
	}

	if err := s.tmpl.RenderDocument(w, data); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
