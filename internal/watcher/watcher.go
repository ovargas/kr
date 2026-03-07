package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors a directory tree for .md file changes with debouncing.
type Watcher struct {
	fsw     *fsnotify.Watcher
	changes chan struct{}
}

// New creates a Watcher that monitors rootPath and immediate subdirectories.
func New(rootPath string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Watch root directory
	if err := fsw.Add(rootPath); err != nil {
		fsw.Close()
		return nil, err
	}

	// Watch immediate subdirectories
	entries, err := os.ReadDir(rootPath)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				fsw.Add(filepath.Join(rootPath, e.Name()))
			}
		}
	}

	w := &Watcher{
		fsw:     fsw,
		changes: make(chan struct{}, 1),
	}

	go w.loop()
	return w, nil
}

// Changes returns a channel that receives a signal when .md files change.
func (w *Watcher) Changes() <-chan struct{} {
	return w.changes
}

// Close shuts down the file watcher.
func (w *Watcher) Close() error {
	return w.fsw.Close()
}

func (w *Watcher) loop() {
	var timer *time.Timer
	var timerC <-chan time.Time

	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}

			// Watch new directories
			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					w.fsw.Add(event.Name)
				}
			}

			// Only trigger on .md file changes
			if !strings.HasSuffix(event.Name, ".md") {
				continue
			}

			// Debounce: reset timer on each event
			if timer == nil {
				timer = time.NewTimer(100 * time.Millisecond)
				timerC = timer.C
			} else {
				timer.Reset(100 * time.Millisecond)
			}

		case <-timerC:
			// Debounce window expired — send notification
			select {
			case w.changes <- struct{}{}:
			default:
				// Channel already has a pending notification
			}
			timer = nil
			timerC = nil

		case _, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
		}
	}
}
