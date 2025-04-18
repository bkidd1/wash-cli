package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Note represents a single note
type Note struct {
	Path    string
	Content string
}

// NewNote creates a new note with the given content
func NewNote(content string) (*Note, error) {
	// Create notes directory if it doesn't exist
	dir := filepath.Join(os.Getenv("HOME"), ".wash-notes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create notes directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("note-%s.md", time.Now().Format("2006-01-02-15-04-05"))
	path := filepath.Join(dir, filename)

	// Create the note file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to create note file: %w", err)
	}

	return &Note{
		Path:    path,
		Content: content,
	}, nil
}

// AppendToNote appends content to an existing note
func AppendToNote(path string, content string) error {
	// Read existing content
	existingContent, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read note file: %w", err)
	}

	// Append new content with a separator
	newContent := string(existingContent) + "\n\n---\n\n" + content

	// Write back to file
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update note file: %w", err)
	}

	return nil
}

// ListNotes returns a list of all notes in the notes directory
func ListNotes() ([]string, error) {
	dir := filepath.Join(os.Getenv("HOME"), ".wash-notes")
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read notes directory: %w", err)
	}

	var notes []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			notes = append(notes, filepath.Join(dir, file.Name()))
		}
	}

	return notes, nil
}
