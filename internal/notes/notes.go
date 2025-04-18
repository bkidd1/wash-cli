package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	dir := filepath.Join(os.Getenv("HOME"), ".wash", "notes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create notes directory: %w", err)
	}

	// Generate filename with timestamp and issue type
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	// Extract issue type from content if present
	issueType := "general"
	if strings.HasPrefix(content, "# ISSUE:") {
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			issueType = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(lines[0], "# ISSUE:")))
			// Clean up issue type for filename
			issueType = strings.ReplaceAll(issueType, " ", "-")
			issueType = strings.ReplaceAll(issueType, "/", "-")
		}
	}
	filename := fmt.Sprintf("note-%s-%s.md", issueType, timestamp)
	path := filepath.Join(dir, filename)

	// Format the content with proper markdown structure
	formattedContent := fmt.Sprintf("# Wash Analysis Note\n*Generated on %s*\n\n%s",
		time.Now().Format("2006-01-02 15:04:05"),
		content)

	// Create the note file
	if err := os.WriteFile(path, []byte(formattedContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to create note file: %w", err)
	}

	return &Note{
		Path:    path,
		Content: formattedContent,
	}, nil
}

// AppendToNote appends content to an existing note
func AppendToNote(path string, content string) error {
	// Read existing content
	existingContent, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read note file: %w", err)
	}

	// Format the new content with a clear separator and timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	separator := fmt.Sprintf("\n\n---\n\n## Update at %s\n\n", timestamp)

	// Append new content with separator
	newContent := string(existingContent) + separator + content

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
