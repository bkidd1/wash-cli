package changetracker

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bkidd1/wash-cli/internal/analyzer"
	"github.com/bkidd1/wash-cli/internal/notes"
	"github.com/fsnotify/fsnotify"
)

// ChangeType represents the type of code change
type ChangeType string

const (
	ChangeTypeFeature  ChangeType = "feature"
	ChangeTypeBugfix   ChangeType = "bugfix"
	ChangeTypeRefactor ChangeType = "refactor"
	ChangeTypeConfig   ChangeType = "config"
	ChangeTypeOther    ChangeType = "other"
)

// CodeChange represents a significant change in code
type CodeChange struct {
	Timestamp    time.Time  `json:"timestamp"`
	ChangeType   ChangeType `json:"change_type"`
	Files        []string   `json:"files"`
	Description  string     `json:"description"`
	Issues       []string   `json:"issues,omitempty"`
	Alternatives []string   `json:"alternatives,omitempty"`
	GitInfo      *GitInfo   `json:"git_info,omitempty"`
}

// GitInfo contains Git-specific information about a change
type GitInfo struct {
	CommitHash string `json:"commit_hash"`
	Branch     string `json:"branch"`
	Author     string `json:"author"`
	Message    string `json:"message"`
}

// ChangeTracker interface defines methods for tracking code changes
type ChangeTracker interface {
	Start() error
	Stop() error
	GetChanges() ([]CodeChange, error)
}

// GitTracker implements ChangeTracker for Git projects
type GitTracker struct {
	projectPath string
	notes       *notes.NotesManager
	analyzer    *analyzer.Analyzer
}

// EventTracker implements ChangeTracker for non-Git projects
type EventTracker struct {
	projectPath  string
	notes        *notes.NotesManager
	analyzer     *analyzer.Analyzer
	watcher      *fsnotify.Watcher
	lastChange   time.Time
	changeBuffer []CodeChange
}

// NewChangeTracker creates an appropriate tracker based on project type
func NewChangeTracker(projectPath string, notes *notes.NotesManager, analyzer *analyzer.Analyzer) (ChangeTracker, error) {
	// Check if the project is a Git repository
	if isGitRepo(projectPath) {
		return &GitTracker{
			projectPath: projectPath,
			notes:       notes,
			analyzer:    analyzer,
		}, nil
	}

	// Create a new file watcher for non-Git projects
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating file watcher: %w", err)
	}

	return &EventTracker{
		projectPath: projectPath,
		notes:       notes,
		analyzer:    analyzer,
		watcher:     watcher,
		lastChange:  time.Now(),
	}, nil
}

// isGitRepo checks if a directory is a Git repository
func isGitRepo(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

// Start begins tracking changes
func (gt *GitTracker) Start() error {
	// Set up Git hooks to track changes
	// This would be implemented in a separate method
	return nil
}

// Start begins tracking changes for non-Git projects
func (et *EventTracker) Start() error {
	// Add the project directory to the watcher
	if err := et.watcher.Add(et.projectPath); err != nil {
		return fmt.Errorf("error adding directory to watcher: %w", err)
	}

	// Start watching for changes
	go et.watchLoop()

	return nil
}

// watchLoop handles file system events for non-Git projects
func (et *EventTracker) watchLoop() {
	for {
		select {
		case event, ok := <-et.watcher.Events:
			if !ok {
				return
			}
			et.handleEvent(event)
		case err, ok := <-et.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("watcher error: %v\n", err)
		}
	}
}

// handleEvent processes file system events
func (et *EventTracker) handleEvent(event fsnotify.Event) {
	// Only process write events
	if event.Op&fsnotify.Write != fsnotify.Write {
		return
	}

	// Check if enough time has passed since the last change
	// This prevents recording every single file save
	if time.Since(et.lastChange) < 5*time.Second {
		return
	}

	// Create a new code change
	change := CodeChange{
		Timestamp:  time.Now(),
		ChangeType: determineChangeType(event.Name),
		Files:      []string{event.Name},
	}

	// Analyze the change
	analysis, err := et.analyzer.AnalyzeFile(context.Background(), event.Name)
	if err != nil {
		fmt.Printf("error analyzing file %s: %v\n", event.Name, err)
		analysis = "Error during analysis: " + err.Error()
	}
	change.Description = analysis

	// Save the change
	et.changeBuffer = append(et.changeBuffer, change)
	et.lastChange = time.Now()

	// Create and save an interaction
	interaction := &notes.Interaction{
		Timestamp:   time.Now(),
		ProjectName: filepath.Base(et.projectPath),
		Type:        notes.InteractionTypeCode,
		Content: struct {
			UserInput   string             `json:"user_input,omitempty"`
			AIResponse  string             `json:"ai_response,omitempty"`
			CodeChanges []notes.CodeChange `json:"code_changes,omitempty"`
		}{
			CodeChanges: []notes.CodeChange{
				{
					File:        event.Name,
					Description: analysis,
				},
			},
		},
	}

	if err := et.notes.SaveInteraction(interaction); err != nil {
		fmt.Printf("error saving interaction: %v\n", err)
	}
}

// determineChangeType determines the type of change based on the file
func determineChangeType(filePath string) ChangeType {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h":
		return ChangeTypeFeature
	case ".json", ".yaml", ".yml", ".toml", ".ini":
		return ChangeTypeConfig
	default:
		return ChangeTypeOther
	}
}

// Stop stops tracking changes
func (gt *GitTracker) Stop() error {
	// Clean up Git hooks
	return nil
}

// Stop stops tracking changes for non-Git projects
func (et *EventTracker) Stop() error {
	return et.watcher.Close()
}

// GetChanges returns all tracked changes
func (gt *GitTracker) GetChanges() ([]CodeChange, error) {
	// Implement Git-specific change retrieval
	return nil, nil
}

// GetChanges returns all tracked changes for non-Git projects
func (et *EventTracker) GetChanges() ([]CodeChange, error) {
	return et.changeBuffer, nil
}
