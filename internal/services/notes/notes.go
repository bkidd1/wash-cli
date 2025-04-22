package notes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// InteractionType represents different types of interactions
type InteractionType string

const (
	InteractionTypeChat     InteractionType = "chat"     // User-AI agent conversations
	InteractionTypeCode     InteractionType = "code"     // Code changes and modifications
	InteractionTypeAnalysis InteractionType = "analysis" // Analysis results and insights
	InteractionTypeDecision InteractionType = "decision" // Key decisions and rationale
	InteractionTypeError    InteractionType = "error"    // Errors and their resolutions
)

// Priority represents the priority level of an interaction
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Status represents the status of an interaction
type Status string

const (
	StatusOpen     Status = "open"
	StatusResolved Status = "resolved"
	StatusArchived Status = "archived"
)

// CodeChange represents a high-level change in code
type CodeChange struct {
	BaseRecord
	File            string   `json:"file"`                   // The file that was modified
	OldContent      string   `json:"old_content"`            // Previous content of the file
	NewContent      string   `json:"new_content"`            // New content of the file
	Description     string   `json:"description"`            // High-level description of the change
	PotentialIssues []string `json:"issues,omitempty"`       // Potential issues to watch out for
	Alternatives    []string `json:"alternatives,omitempty"` // Alternative approaches considered
}

// SetBaseRecord implements the BaseRecord interface
func (cc *CodeChange) SetBaseRecord(br BaseRecord) {
	cc.BaseRecord = br
}

// GetTimestamp implements the BaseRecord interface
func (cc *CodeChange) GetTimestamp() time.Time {
	return cc.Timestamp
}

// Analysis represents the analysis of an interaction
type Analysis struct {
	Suggestions           []string `json:"suggestions,omitempty"`
	PotentialIssues       []string `json:"potential_issues,omitempty"`
	AlternativeApproaches []string `json:"alternative_approaches,omitempty"`
}

// Interaction represents a single interaction between user and AI
type Interaction struct {
	BaseRecord
	Timestamp   time.Time `json:"timestamp"`
	ProjectName string    `json:"project_name"`
	ProjectGoal string    `json:"project_goal"`
	Context     struct {
		CurrentState string   `json:"current_state"`
		FilesChanged []string `json:"files_changed,omitempty"`
	} `json:"context"`
	Analysis struct {
		CurrentApproach string   `json:"current_approach"`
		Issues          []string `json:"issues,omitempty"`
		Solutions       []string `json:"solutions,omitempty"`
		BestPractices   []string `json:"best_practices,omitempty"`
	} `json:"analysis"`
	Metadata struct {
		Tags     []string `json:"tags,omitempty"`
		Priority Priority `json:"priority,omitempty"`
		Status   Status   `json:"status,omitempty"`
	} `json:"metadata"`
}

// SetBaseRecord implements the BaseRecord interface
func (i *Interaction) SetBaseRecord(br BaseRecord) {
	i.BaseRecord = br
}

// GetTimestamp implements the BaseRecord interface
func (i *Interaction) GetTimestamp() time.Time {
	return i.Timestamp
}

// RememberNote represents a user-created note from wash remember
type RememberNote struct {
	BaseRecord
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

// SetBaseRecord implements the BaseRecord interface
func (n *RememberNote) SetBaseRecord(br BaseRecord) {
	n.BaseRecord = br
}

// GetTimestamp implements the BaseRecord interface
func (n *RememberNote) GetTimestamp() time.Time {
	return n.Timestamp
}

// MonitorNote represents a note from wash monitor
type MonitorNote struct {
	BaseRecord
	Timestamp   time.Time `json:"timestamp"`
	ProjectName string    `json:"project_name"`
	ProjectGoal string    `json:"project_goal"`
	Context     struct {
		CurrentState string   `json:"current_state"`
		FilesChanged []string `json:"files_changed,omitempty"`
	} `json:"context"`
	Analysis struct {
		CurrentApproach string   `json:"current_approach"`
		Issues          []string `json:"issues,omitempty"`
		Solutions       []string `json:"solutions,omitempty"`
		BestPractices   []string `json:"best_practices,omitempty"`
	} `json:"analysis"`
	Metadata struct {
		Tags     []string `json:"tags,omitempty"`
		Priority Priority `json:"priority,omitempty"`
		Status   Status   `json:"status,omitempty"`
	} `json:"metadata"`
}

// SetBaseRecord implements the BaseRecord interface
func (n *MonitorNote) SetBaseRecord(br BaseRecord) {
	n.BaseRecord = br
}

// GetTimestamp implements the BaseRecord interface
func (n *MonitorNote) GetTimestamp() time.Time {
	return n.Timestamp
}

// FileNote represents a note from wash file
type FileNote struct {
	BaseRecord
	File            string   `json:"file"`
	Analysis        string   `json:"analysis"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// SetBaseRecord implements the BaseRecord interface
func (n *FileNote) SetBaseRecord(br BaseRecord) {
	n.BaseRecord = br
}

// GetTimestamp implements the BaseRecord interface
func (n *FileNote) GetTimestamp() time.Time {
	return n.Timestamp
}

// ProjectNote represents a note from wash project
type ProjectNote struct {
	BaseRecord
	ProjectName string `json:"project_name"`
	Structure   struct {
		Files       []string `json:"files"`
		Directories []string `json:"directories"`
		Issues      []string `json:"issues,omitempty"`
	} `json:"structure"`
	Analysis struct {
		Architecture    string   `json:"architecture"`
		Dependencies    []string `json:"dependencies"`
		Recommendations []string `json:"recommendations,omitempty"`
	} `json:"analysis"`
}

// SetBaseRecord implements the BaseRecord interface
func (n *ProjectNote) SetBaseRecord(br BaseRecord) {
	n.BaseRecord = br
}

// GetTimestamp implements the BaseRecord interface
func (n *ProjectNote) GetTimestamp() time.Time {
	return n.Timestamp
}

// BugNote represents a note from wash bug
type BugNote struct {
	BaseRecord
	Description      string   `json:"description"`
	StepsToReproduce []string `json:"steps_to_reproduce"`
	ExpectedBehavior string   `json:"expected_behavior"`
	ActualBehavior   string   `json:"actual_behavior"`
	Status           Status   `json:"status"`
	Priority         Priority `json:"priority"`
	Solution         string   `json:"solution,omitempty"`
}

// SetBaseRecord implements the BaseRecord interface
func (n *BugNote) SetBaseRecord(br BaseRecord) {
	n.BaseRecord = br
}

// GetTimestamp implements the BaseRecord interface
func (n *BugNote) GetTimestamp() time.Time {
	return n.Timestamp
}

// Manager handles note storage and retrieval
type Manager struct {
	// TODO: Add storage backend (e.g., database, file system)
}

// NewManager creates a new notes manager
func NewManager() *Manager {
	return &Manager{}
}

// Save stores a note
func (m *Manager) Save(note *RememberNote) error {
	// TODO: Implement actual storage
	return nil
}

// Get retrieves notes matching the given criteria
func (m *Manager) Get(filter map[string]interface{}) ([]*RememberNote, error) {
	// TODO: Implement actual retrieval
	return nil, nil
}

// NotesManager handles all Wash notes operations
type NotesManager struct {
	baseDir string
}

// NewNotesManager creates a new NotesManager instance
func NewNotesManager() (*NotesManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".wash")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating .wash directory: %w", err)
	}

	// Create necessary subdirectories based on commands/actions
	dirs := []string{
		"changelog",     // Code change history and decisions
		"monitor_notes", // Monitoring and interaction notes
		"analyze",       // Code analysis results
		"config",        // User configuration and preferences
		"errors",        // Error tracking and debugging
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0755); err != nil {
			return nil, fmt.Errorf("error creating %s directory: %w", dir, err)
		}
	}

	return &NotesManager{baseDir: baseDir}, nil
}

// SaveInteraction saves a new interaction
func (nm *NotesManager) SaveInteraction(interaction *Interaction) error {
	// Create project-specific directory
	projectDir := filepath.Join(nm.baseDir, "projects", interaction.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("error creating project directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("%s.json", interaction.Timestamp.Format("2006-01-02-15-04-05"))
	notesDir := filepath.Join(projectDir, "notes")
	filepath := filepath.Join(notesDir, filename)

	// Create notes directory if it doesn't exist
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return fmt.Errorf("error creating notes directory: %w", err)
	}

	// Save interaction to file
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating interaction file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(interaction); err != nil {
		return fmt.Errorf("error encoding interaction: %w", err)
	}

	return nil
}

// SaveUserNote saves a user-specific note
func (nm *NotesManager) SaveUserNote(username string, note *RememberNote) error {
	userDir := filepath.Join(nm.baseDir, "remember", username)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("error creating user directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("%s_%s.json", note.Timestamp.Format("2006-01-02-15-04-05"), uuid.New().String())
	filepath := filepath.Join(userDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating note file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(note); err != nil {
		return fmt.Errorf("error encoding note: %w", err)
	}

	return nil
}

// LoadInteractions loads all interactions for a project
func (nm *NotesManager) LoadInteractions(projectName string) ([]*Interaction, error) {
	projectDir := filepath.Join(nm.baseDir, "projects", projectName, "notes")

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, nil
	}

	files, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, fmt.Errorf("error reading notes directory: %w", err)
	}

	var interactions []*Interaction
	for _, file := range files {
		// Skip non-JSON files
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filepath := filepath.Join(projectDir, file.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: Could not read file %s: %v\n", file.Name(), err)
			continue
		}

		var interaction Interaction
		if err := json.Unmarshal(data, &interaction); err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: Could not parse JSON in file %s: %v\n", file.Name(), err)
			continue
		}

		interactions = append(interactions, &interaction)
	}

	return interactions, nil
}

// QueryInteractions queries interactions based on criteria
func (nm *NotesManager) QueryInteractions(projectName string, criteria map[string]interface{}) ([]*Interaction, error) {
	interactions, err := nm.LoadInteractions(projectName)
	if err != nil {
		return nil, err
	}

	var filtered []*Interaction
	for _, interaction := range interactions {
		if matchesCriteria(interaction, criteria) {
			filtered = append(filtered, interaction)
		}
	}

	return filtered, nil
}

// matchesCriteria checks if an interaction matches the given criteria
func matchesCriteria(interaction *Interaction, criteria map[string]interface{}) bool {
	for key, value := range criteria {
		switch key {
		case "priority":
			if interaction.Metadata.Priority != value.(Priority) {
				return false
			}
		case "status":
			if interaction.Metadata.Status != value.(Status) {
				return false
			}
		case "tag":
			tag := value.(string)
			found := false
			for _, t := range interaction.Metadata.Tags {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}
