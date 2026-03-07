package server

import (
	"fmt"
	"net/http"
	"sync"
)

// SSEBroker manages SSE client connections and broadcasts events.
type SSEBroker struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

// NewSSEBroker creates a new SSE broker.
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		clients: make(map[chan struct{}]struct{}),
	}
}

// Subscribe registers a new client and returns its channel.
func (b *SSEBroker) Subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a client and closes its channel.
func (b *SSEBroker) Unsubscribe(ch chan struct{}) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

// Broadcast sends a notification to all connected clients.
func (b *SSEBroker) Broadcast() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.broker.Subscribe()
	defer s.broker.Unsubscribe(ch)

	// Send initial connected event
	fmt.Fprint(w, "event: connected\ndata: ok\n\n")
	flusher.Flush()

	for {
		select {
		case <-ch:
			fmt.Fprint(w, "event: change\ndata: reload\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
