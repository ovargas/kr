package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/ovargas/kr/internal/renderer"
	"github.com/ovargas/kr/internal/static"
	"github.com/ovargas/kr/internal/templates"
	"github.com/ovargas/kr/internal/watcher"
)

// Server holds the HTTP server configuration and state.
type Server struct {
	port     int
	rootPath string
	mux      *http.ServeMux
	renderer *renderer.Renderer
	tmpl     *templates.Templates
	watcher  *watcher.Watcher
	broker   *SSEBroker
}

// New creates a new Server with all routes registered.
func New(port int, rootPath string) (*Server, error) {
	tmpl, err := templates.New()
	if err != nil {
		return nil, err
	}

	w, err := watcher.New(rootPath)
	if err != nil {
		return nil, err
	}

	broker := NewSSEBroker()

	mux := http.NewServeMux()
	s := &Server{
		port:     port,
		rootPath: rootPath,
		mux:      mux,
		renderer: renderer.New(),
		tmpl:     tmpl,
		watcher:  w,
		broker:   broker,
	}

	// Forward watcher changes to SSE broker
	go func() {
		for range w.Changes() {
			broker.Broadcast()
		}
	}()

	mux.HandleFunc("GET /{$}", s.handleBacklog)
	mux.HandleFunc("GET /events", s.handleSSE)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))
	mux.HandleFunc("GET /", s.handleCatchAll)

	return s, nil
}

// handleCatchAll dispatches folder and document requests.
// Paths like /{folder}/ go to handleFolder, /{folder}/{file} go to handleDocument.
func (s *Server) handleCatchAll(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	path = strings.TrimSuffix(path, "/")

	parts := strings.SplitN(path, "/", 3)
	switch {
	case len(parts) == 1 && parts[0] != "":
		// /{folder}/ — folder listing
		r.SetPathValue("folder", parts[0])
		s.handleFolder(w, r)
	case len(parts) == 2:
		// /{folder}/{file} — document view
		r.SetPathValue("folder", parts[0])
		r.SetPathValue("file", parts[1])
		s.handleDocument(w, r)
	default:
		http.NotFound(w, r)
	}
}

// Start binds to the configured port and serves HTTP requests.
// It prints the server URL to stdout and blocks until the server exits.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("http://localhost:%d\n", port)

	return http.Serve(listener, s.mux)
}

// Close shuts down the file watcher.
func (s *Server) Close() error {
	if s.watcher != nil {
		return s.watcher.Close()
	}
	return nil
}
