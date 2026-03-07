package server

import (
	"fmt"
	"net"
	"net/http"
)

// Server holds the HTTP server configuration and state.
type Server struct {
	port     int
	rootPath string
	mux      *http.ServeMux
}

// New creates a new Server with a placeholder route.
func New(port int, rootPath string) *Server {
	mux := http.NewServeMux()
	s := &Server{
		port:     port,
		rootPath: rootPath,
		mux:      mux,
	}

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "kr is running")
	})

	return s
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
