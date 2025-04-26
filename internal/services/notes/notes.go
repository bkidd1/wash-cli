package notes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
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

// Analysis represents the analysis of an interaction
type Analysis struct {
	Suggestions           []string `json:"suggestions,omitempty"`
	PotentialIssues       []string `json:"potential_issues,omitempty"`
	AlternativeApproaches []string `json:"alternative_approaches,omitempty"`
}

// MonitorAnalysis represents the analysis of monitor notes
type MonitorAnalysis struct {
	Summary                 string   `json:"summary"`
	PotentialIssues         []string `json:"potential_issues"`
	OptimizationSuggestions []string `json:"optimization_suggestions"`
	FilesChanged            []string `json:"files_changed"`
}

// Interaction represents a single interaction between user and AI
type Interaction struct {
	Timestamp   time.Time `json:"timestamp"`
	ProjectName string    `json:"project_name"`
	ProjectGoal string    `json:"project_goal"`
	Context     struct {
		CurrentState string   `json:"current_state"`
		FilesChanged []string `json:"files_changed,omitempty"`
	} `json:"context"`
	Analysis struct {
		CurrentApproach       string   `json:"current_approach"`
		AlternativeApproaches []string `json:"alternative_approaches,omitempty"`
	} `json:"analysis"`
	Metadata struct {
		Tags     []string `json:"tags,omitempty"`
		Priority Priority `json:"priority,omitempty"`
		Status   Status   `json:"status,omitempty"`
	} `json:"metadata"`
}

// MonitorNote represents a note from wash monitor
type MonitorNote struct {
	Timestamp   time.Time `json:"timestamp"`
	ProjectName string    `json:"project_name"`
	Interaction struct {
		UserRequest string   `json:"user_request"`
		AIAction    string   `json:"ai_action"`
		Context     string   `json:"context"`
		CodeChanges []string `json:"code_changes"`
	} `json:"interaction"`
}

// ProjectProgressNote represents significant project progress and milestones
type ProjectProgressNote struct {
	Timestamp   time.Time `json:"timestamp"`
	ID          string    `json:"id"`
	ProjectName string    `json:"project_name"`
	Type        string    `json:"type"` // e.g., "milestone", "architecture", "feature", "refactor"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Changes     struct {
		FilesModified []string `json:"files_modified,omitempty"`
		FilesAdded    []string `json:"files_added,omitempty"`
		FilesDeleted  []string `json:"files_deleted,omitempty"`
	} `json:"changes"`
	Impact struct {
		Scope         string   `json:"scope"` // e.g., "local", "module", "project-wide"
		AffectedAreas []string `json:"affected_areas,omitempty"`
		RiskLevel     string   `json:"risk_level"` // e.g., "low", "medium", "high"
	} `json:"impact"`
	Metadata struct {
		Tags     []string `json:"tags,omitempty"`
		Priority Priority `json:"priority,omitempty"`
		Status   Status   `json:"status,omitempty"`
	} `json:"metadata"`
}

// RememberNote represents a user-created note from wash remember
type RememberNote struct {
	Timestamp time.Time              `json:"timestamp"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
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
		"progress",      // Project progress notes
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

// SaveProjectProgress saves a project progress note
func (nm *NotesManager) SaveProjectProgress(note *ProjectProgressNote) error {
	note.Timestamp = time.Now()
	note.ID = uuid.New().String()

	// Create the progress directory if it doesn't exist
	progressDir := filepath.Join(nm.baseDir, "progress")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		return fmt.Errorf("error creating progress directory: %w", err)
	}

	// Create a file for the note
	noteFile := filepath.Join(progressDir, fmt.Sprintf("%s_%s.json", note.ProjectName, note.ID))
	data, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling note: %w", err)
	}

	if err := os.WriteFile(noteFile, data, 0644); err != nil {
		return fmt.Errorf("error writing note file: %w", err)
	}

	return nil
}

// LoadProjectProgress loads all project progress notes for a given project
func (nm *NotesManager) LoadProjectProgress(projectName string) ([]*ProjectProgressNote, error) {
	progressDir := filepath.Join(nm.baseDir, "progress")
	files, err := os.ReadDir(progressDir)
	if err != nil {
		return nil, fmt.Errorf("error reading progress directory: %w", err)
	}

	var notes []*ProjectProgressNote
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), projectName+"_") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(progressDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("error reading note file: %w", err)
		}

		var note ProjectProgressNote
		if err := json.Unmarshal(data, &note); err != nil {
			return nil, fmt.Errorf("error unmarshaling note: %w", err)
		}

		notes = append(notes, &note)
	}

	return notes, nil
}

// QueryProjectProgress queries project progress notes based on criteria
func (nm *NotesManager) QueryProjectProgress(projectName string, criteria map[string]interface{}) ([]*ProjectProgressNote, error) {
	notes, err := nm.LoadProjectProgress(projectName)
	if err != nil {
		return nil, err
	}

	var filteredNotes []*ProjectProgressNote
	for _, note := range notes {
		if matchesProgressCriteria(note, criteria) {
			filteredNotes = append(filteredNotes, note)
		}
	}

	return filteredNotes, nil
}

// matchesProgressCriteria checks if a note matches the given criteria
func matchesProgressCriteria(note *ProjectProgressNote, criteria map[string]interface{}) bool {
	for key, value := range criteria {
		switch key {
		case "type":
			if note.Type != value.(string) {
				return false
			}
		case "priority":
			if note.Metadata.Priority != value.(Priority) {
				return false
			}
		case "status":
			if note.Metadata.Status != value.(Status) {
				return false
			}
		case "tag":
			tag := value.(string)
			found := false
			for _, t := range note.Metadata.Tags {
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

// GenerateProgressFromMonitor generates a progress note from recent monitor data
func (nm *NotesManager) GenerateProgressFromMonitor(projectName string, duration time.Duration) (*ProjectProgressNote, error) {
	// Get recent monitor notes
	monitorDir := filepath.Join(nm.baseDir, "monitor_notes", projectName)
	files, err := os.ReadDir(monitorDir)
	if err != nil {
		return nil, fmt.Errorf("error reading monitor directory: %w", err)
	}

	// Get the cutoff time
	cutoffTime := time.Now().Add(-duration)

	var recentNotes []*MonitorNote
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(monitorDir, file.Name()))
		if err != nil {
			continue
		}

		var note MonitorNote
		if err := json.Unmarshal(data, &note); err != nil {
			continue
		}

		if note.Timestamp.After(cutoffTime) {
			recentNotes = append(recentNotes, &note)
		}
	}

	if len(recentNotes) == 0 {
		return nil, fmt.Errorf("no monitor notes found in the last %v", duration)
	}

	// Create the progress note
	progressNote := &ProjectProgressNote{
		Timestamp:   time.Now(),
		ID:          uuid.New().String(),
		ProjectName: projectName,
		Type:        "summary",
		Title:       fmt.Sprintf("5-Minute Summary"),
	}

	// Format monitor notes for API analysis
	var monitorData strings.Builder
	for _, note := range recentNotes {
		monitorData.WriteString(fmt.Sprintf("Time: %s\n", note.Timestamp.Format("2006-01-02 15:04:05")))
		monitorData.WriteString(fmt.Sprintf("User Request: %s\n", note.Interaction.UserRequest))
		monitorData.WriteString(fmt.Sprintf("AI Action: %s\n", note.Interaction.AIAction))
		monitorData.WriteString(fmt.Sprintf("Context: %s\n", note.Interaction.Context))
		if len(note.Interaction.CodeChanges) > 0 {
			monitorData.WriteString(fmt.Sprintf("Code Changes: %s\n", strings.Join(note.Interaction.CodeChanges, ", ")))
		}
		monitorData.WriteString("\n")
	}

	// Load config to get API key
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create API client with config key
	client := openai.NewClient(cfg.OpenAIKey)

	// Create the analysis prompt
	prompt := `You are analyzing a series of development interactions between a user and an AI coding assistant.
Your task is to analyze these interactions and provide a comprehensive summary and analysis. 
The user wants to know if they made any poor suggestions in the chat/misguided AI actions. It's your job to identify these miscommunications.
Also look out for any AI actions that seemed to take a sub-optimal approach for the user's overall goal.

Monitor Notes:
` + monitorData.String() + `

Please analyze these interactions and provide:
1. A concise paragraph summarizing the main development activities and progress
2. Potential errors or issues that were introduced through bad prompts in the chat/misguided AI actions
3. Suggestions for alternative, more optimized approaches to:
   - Errors caused by bad prompts in the chat/miscommunications
   - Project structure and organization
   - Features and functionality
   - Build process and tools
   - Code quality and maintainability
4. A list of all files that were modified

Format your response as a JSON object with the following structure:
{
    "summary": "paragraph summarizing main activities and progress",
    "potential_issues": ["list of potential issues or errors"],
    "optimization_suggestions": ["list of suggestions for better approaches"],
    "files_changed": ["list of modified files"]
}`

	// Call the API
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting analysis: %w", err)
	}

	// Parse the response
	var analysis MonitorAnalysis
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
		return nil, fmt.Errorf("error parsing analysis: %w", err)
	}

	// Update the progress note with the analysis
	progressNote.Description = fmt.Sprintf("Summary:\n%s\n\nPotential Issues:\n%s\n\nOptimization Suggestions:\n%s\n\nFiles Changed:\n%s",
		analysis.Summary,
		strings.Join(analysis.PotentialIssues, "\n"),
		strings.Join(analysis.OptimizationSuggestions, "\n"),
		strings.Join(analysis.FilesChanged, "\n"),
	)

	// Set impact assessment
	progressNote.Impact.Scope = "project-wide"
	if len(analysis.PotentialIssues) > 0 {
		progressNote.Impact.RiskLevel = "medium"
	} else {
		progressNote.Impact.RiskLevel = "low"
	}

	// Set metadata
	progressNote.Metadata.Priority = PriorityLow
	progressNote.Metadata.Status = StatusOpen
	progressNote.Metadata.Tags = []string{"summary", "auto_generated"}

	return progressNote, nil
}

// GetMonitorNotesDir returns the path to the monitor notes directory for a project
func (nm *NotesManager) GetMonitorNotesDir(projectName string) string {
	return filepath.Join(nm.baseDir, "monitor_notes", projectName)
}

// GetUserNotes retrieves all remember notes for a specific user and project
func (nm *NotesManager) GetUserNotes(username string, projectName string) ([]*RememberNote, error) {
	userDir := filepath.Join(nm.baseDir, "remember", username)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return nil, nil
	}

	files, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("error reading user directory: %w", err)
	}

	var notes []*RememberNote
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(userDir, file.Name()))
		if err != nil {
			continue
		}

		var note RememberNote
		if err := json.Unmarshal(data, &note); err != nil {
			continue
		}

		// Check if the note belongs to the specified project
		if project, ok := note.Metadata["project"].(string); ok && project == projectName {
			notes = append(notes, &note)
		}
	}

	return notes, nil
}

// SaveMonitorNote saves a monitor note for a project
func (nm *NotesManager) SaveMonitorNote(projectName string, note *MonitorNote) error {
	// Create project-specific directory
	projectDir := filepath.Join(nm.baseDir, "monitor_notes", projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("error creating project directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("%s.json", note.Timestamp.Format("2006-01-02-15-04-05"))
	filepath := filepath.Join(projectDir, filename)

	// Save note to file
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

// GetProgressNotes retrieves all progress notes for a specific project
func (nm *NotesManager) GetProgressNotes(projectName string) ([]*ProjectProgressNote, error) {
	progressDir := filepath.Join(nm.baseDir, "progress")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating progress directory: %w", err)
	}

	entries, err := os.ReadDir(progressDir)
	if err != nil {
		return nil, fmt.Errorf("error reading progress directory: %w", err)
	}

	var notes []*ProjectProgressNote
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Check if the file belongs to the specified project
		if !strings.HasPrefix(entry.Name(), projectName+"_") {
			continue
		}

		filePath := filepath.Join(progressDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading progress note file %s: %w", entry.Name(), err)
		}

		var note ProjectProgressNote
		if err := json.Unmarshal(data, &note); err != nil {
			return nil, fmt.Errorf("error unmarshaling progress note from %s: %w", entry.Name(), err)
		}

		notes = append(notes, &note)
	}

	return notes, nil
}
