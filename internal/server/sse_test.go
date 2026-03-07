package server

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSEConnection(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}

	scanner := bufio.NewScanner(resp.Body)
	var lines []string
	done := make(chan struct{})
	go func() {
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
			if strings.Contains(scanner.Text(), "data: ok") {
				close(done)
				return
			}
		}
	}()

	select {
	case <-done:
		// Check for connected event
		found := false
		for _, l := range lines {
			if strings.Contains(l, "event: connected") {
				found = true
				break
			}
		}
		if !found {
			t.Error("missing 'event: connected' in SSE stream")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for SSE connected event")
	}
}

func TestSSEBroadcast(t *testing.T) {
	srv, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Wait for connected event
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "data: ok") {
			break
		}
	}

	// Broadcast a change
	srv.broker.Broadcast()

	// Read change event
	done := make(chan bool)
	go func() {
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "data: reload") {
				done <- true
				return
			}
		}
		done <- false
	}()

	select {
	case ok := <-done:
		if !ok {
			t.Error("did not receive change event")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for change event")
	}
}

func TestSSEMultipleClients(t *testing.T) {
	broker := NewSSEBroker()
	ch1 := broker.Subscribe()
	ch2 := broker.Subscribe()

	broker.Broadcast()

	select {
	case <-ch1:
	case <-time.After(100 * time.Millisecond):
		t.Error("client 1 did not receive broadcast")
	}

	select {
	case <-ch2:
	case <-time.After(100 * time.Millisecond):
		t.Error("client 2 did not receive broadcast")
	}

	broker.Unsubscribe(ch1)
	broker.Unsubscribe(ch2)
}

func TestSSEClientDisconnect(t *testing.T) {
	broker := NewSSEBroker()
	ch := broker.Subscribe()

	broker.mu.Lock()
	count := len(broker.clients)
	broker.mu.Unlock()
	if count != 1 {
		t.Errorf("clients = %d, want 1", count)
	}

	broker.Unsubscribe(ch)

	broker.mu.Lock()
	count = len(broker.clients)
	broker.mu.Unlock()
	if count != 0 {
		t.Errorf("clients = %d after unsubscribe, want 0", count)
	}
}
