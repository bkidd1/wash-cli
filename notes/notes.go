package notes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// NoteType represents different types of Wash notes
type NoteType string

const (
	NoteTypeUser      NoteType = "user"
	NoteTypeChat      NoteType = "chat"
	NoteTypeChangelog NoteType = "changelog"
	NoteTypeProject   NoteType = "project"
)

// Note represents a unified note structure
type Note struct {
	Type        NoteType               `json:"type"`
	Content     string                 `json:"content"`
	Timestamp   time.Time              `json:"timestamp"`
	ProjectName string                 `json:"project_name,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
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

	return &NotesManager{baseDir: baseDir}, nil
}

// SaveNote saves a note of any type
func (nm *NotesManager) SaveNote(note Note) error {
	// Create project directory if it doesn't exist
	projectDir := filepath.Join(nm.baseDir, "projects", note.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("error creating project directory: %w", err)
	}

	// Create type-specific directory
	typeDir := filepath.Join(projectDir, string(note.Type))
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		return fmt.Errorf("error creating type directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("%s_%d.json", string(note.Type), note.Timestamp.Unix())
	filepath := filepath.Join(typeDir, filename)

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

// LoadNotes loads all notes for a project
func (nm *NotesManager) LoadNotes(projectName string) ([]Note, error) {
	projectDir := filepath.Join(nm.baseDir, "projects", projectName)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("project directory does not exist: %w", err)
	}

	var notes []Note

	// Load notes from all type directories
	types := []NoteType{NoteTypeUser, NoteTypeChat, NoteTypeChangelog, NoteTypeProject}
	for _, noteType := range types {
		typeDir := filepath.Join(projectDir, string(noteType))
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		files, err := os.ReadDir(typeDir)
		if err != nil {
			return nil, fmt.Errorf("error reading %s directory: %w", noteType, err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filepath := filepath.Join(typeDir, file.Name())
			content, err := os.ReadFile(filepath)
			if err != nil {
				return nil, fmt.Errorf("error reading note file: %w", err)
			}

			var note Note
			if err := json.Unmarshal(content, &note); err != nil {
				return nil, fmt.Errorf("error decoding note: %w", err)
			}

			notes = append(notes, note)
		}
	}

	return notes, nil
}

// GetNotesByType loads notes of a specific type for a project
func (nm *NotesManager) GetNotesByType(projectName string, noteType NoteType) ([]Note, error) {
	notes, err := nm.LoadNotes(projectName)
	if err != nil {
		return nil, err
	}

	var filteredNotes []Note
	for _, note := range notes {
		if note.Type == noteType {
			filteredNotes = append(filteredNotes, note)
		}
	}

	return filteredNotes, nil
}

// FormatNotesForAnalysis formats all notes for use in AI analysis
func (nm *NotesManager) FormatNotesForAnalysis(projectName string) (string, error) {
	notes, err := nm.LoadNotes(projectName)
	if err != nil {
		return "", err
	}

	if len(notes) == 0 {
		return "", nil
	}

	formatted := "Project History and Context:\n\n"
	for _, note := range notes {
		formatted += fmt.Sprintf("[%s] %s\n%s\n\n",
			note.Type,
			note.Timestamp.Format("2006-01-02 15:04:05"),
			note.Content)
	}

	return formatted, nil
}

// Cleanup deletes all existing notes and directories
func (nm *NotesManager) Cleanup() error {
	// Delete the entire .wash directory
	if err := os.RemoveAll(nm.baseDir); err != nil {
		return fmt.Errorf("error deleting .wash directory: %w", err)
	}

	// Recreate the base directory
	if err := os.MkdirAll(nm.baseDir, 0755); err != nil {
		return fmt.Errorf("error recreating .wash directory: %w", err)
	}

	return nil
}
