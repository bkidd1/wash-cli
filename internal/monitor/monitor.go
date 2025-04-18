package monitor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Monitor represents a file system monitor
type Monitor struct {
	watcher *fsnotify.Watcher
	paths   []string
	events  chan Event
	done    chan struct{}
}

// Event represents a file system event
type Event struct {
	Path      string
	Type      string
	Timestamp time.Time
}

// NewMonitor creates a new file system monitor
func NewMonitor(paths []string) (*Monitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &Monitor{
		watcher: watcher,
		paths:   paths,
		events:  make(chan Event, 100),
		done:    make(chan struct{}),
	}, nil
}

// Start begins monitoring the specified paths
func (m *Monitor) Start() error {
	// Add paths to watcher
	for _, path := range m.paths {
		// If path is a directory, watch it recursively
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return m.watcher.Add(path)
				}
				return nil
			}); err != nil {
				return fmt.Errorf("failed to add directory %s to watcher: %w", path, err)
			}
		} else {
			if err := m.watcher.Add(path); err != nil {
				return fmt.Errorf("failed to add path %s to watcher: %w", path, err)
			}
		}
	}

	// Start watching for events
	go func() {
		for {
			select {
			case event, ok := <-m.watcher.Events:
				if !ok {
					return
				}
				m.handleEvent(event)
			case err, ok := <-m.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("error watching files: %v", err)
			case <-m.done:
				return
			}
		}
	}()

	return nil
}

// Stop stops the file system monitor
func (m *Monitor) Stop() error {
	close(m.done)
	return m.watcher.Close()
}

// Events returns a channel that receives file system events
func (m *Monitor) Events() <-chan Event {
	return m.events
}

// handleEvent processes file system events
func (m *Monitor) handleEvent(event fsnotify.Event) {
	// Skip directories and hidden files
	if strings.HasPrefix(filepath.Base(event.Name), ".") {
		return
	}

	var eventType string
	switch event.Op {
	case fsnotify.Create:
		eventType = "create"
	case fsnotify.Write:
		eventType = "write"
	case fsnotify.Remove:
		eventType = "remove"
	case fsnotify.Rename:
		eventType = "rename"
	case fsnotify.Chmod:
		eventType = "chmod"
	}

	m.events <- Event{
		Path:      event.Name,
		Type:      eventType,
		Timestamp: time.Now(),
	}
}
