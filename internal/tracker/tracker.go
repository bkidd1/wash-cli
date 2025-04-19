package tracker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Change represents a file change
type Change struct {
	FilePath    string
	OldContent  string
	NewContent  string
	Timestamp   time.Time
	Description string
}

// Error represents a project error
type Error struct {
	Message           string
	FilePath          string
	LineNumber        int
	Timestamp         time.Time
	StackTrace        string
	RelatedDecisionID string
}

// Decision represents a key decision point in the project
type Decision struct {
	ID              string
	Timestamp       time.Time
	OriginalAsk     string
	Implementation  string
	Changes         []Change
	PotentialIssues []string
	Alternatives    []Alternative
}

// Alternative represents a better approach to a decision
type Alternative struct {
	Description    string
	Benefits       []string
	Implementation string
	CodeExample    string
}

// ProjectState tracks the current state of the project
type ProjectState struct {
	ProjectPath      string
	CurrentFiles     map[string]string
	RecentChanges    []Change
	ActiveErrors     []Error
	DecisionPoints   []Decision
	AlternativePaths []Alternative
	LastUpdated      time.Time
}

// NewProjectState creates a new project state tracker
func NewProjectState(projectPath string) (*ProjectState, error) {
	// Create project-specific state directory
	stateDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", filepath.Base(projectPath), "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	// Create .gitignore in state directory
	gitignorePath := filepath.Join(stateDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to create .gitignore: %w", err)
	}

	// Try to load existing state
	statePath := filepath.Join(stateDir, "state.json")
	if _, err := os.Stat(statePath); err == nil {
		// State file exists, try to load it
		data, err := os.ReadFile(statePath)
		if err == nil {
			var state ProjectState
			if err := json.Unmarshal(data, &state); err == nil {
				return &state, nil
			}
		}
	}

	// Create new state if loading fails or file doesn't exist
	state := &ProjectState{
		ProjectPath:  projectPath,
		CurrentFiles: make(map[string]string),
		LastUpdated:  time.Now(),
	}

	// Save initial state
	if err := state.saveState(); err != nil {
		return nil, fmt.Errorf("failed to save initial state: %w", err)
	}

	return state, nil
}

// TrackChange records a file change
func (ps *ProjectState) TrackChange(change Change) error {
	ps.RecentChanges = append(ps.RecentChanges, change)
	ps.CurrentFiles[change.FilePath] = change.NewContent
	ps.LastUpdated = time.Now()
	return ps.saveState()
}

// TrackError records a project error
func (ps *ProjectState) TrackError(err Error) error {
	ps.ActiveErrors = append(ps.ActiveErrors, err)
	ps.LastUpdated = time.Now()
	return ps.saveState()
}

// TrackDecision records a key decision point
func (ps *ProjectState) TrackDecision(decision Decision) error {
	ps.DecisionPoints = append(ps.DecisionPoints, decision)
	ps.LastUpdated = time.Now()
	return ps.saveState()
}

// saveState persists the current project state
func (ps *ProjectState) saveState() error {
	stateDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", filepath.Base(ps.ProjectPath), "state")
	statePath := filepath.Join(stateDir, "state.json")

	// Marshal the state to JSON
	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write the state file
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// loadState loads the project state from disk
func loadState(projectPath string) (*ProjectState, error) {
	stateDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", filepath.Base(projectPath), "state")
	statePath := filepath.Join(stateDir, "state.json")

	// Read the state file
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state: %w", err)
	}

	// Unmarshal the state
	var state ProjectState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// FindRelatedDecisions finds decisions related to a specific error
func (ps *ProjectState) FindRelatedDecisions(err Error) []Decision {
	var related []Decision
	for _, decision := range ps.DecisionPoints {
		// Check if error is related to files changed in this decision
		for _, change := range decision.Changes {
			if change.FilePath == err.FilePath {
				related = append(related, decision)
				break
			}
		}
	}
	return related
}

// GetAlternativePaths returns alternative approaches for a given error
func (ps *ProjectState) GetAlternativePaths(err Error) []Alternative {
	var alternatives []Alternative
	relatedDecisions := ps.FindRelatedDecisions(err)

	for _, decision := range relatedDecisions {
		alternatives = append(alternatives, decision.Alternatives...)
	}

	return alternatives
}
