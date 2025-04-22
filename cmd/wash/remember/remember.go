package remember

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/notes"
	"github.com/spf13/cobra"
)

// Command returns the remember command
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "remember [content]",
		Short: "Save something to remember",
		Long:  `Save something to remember for your project. Items are stored in ~/.wash/remember/[project-name]/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var content string
			if len(args) == 0 {
				// Interactive mode
				fmt.Print("Enter your note: ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				content = strings.TrimSpace(input)
			} else {
				// Command line argument mode
				content = strings.TrimSpace(strings.Join(args, " "))
			}

			// Validate content
			if content == "" {
				return fmt.Errorf("content cannot be empty")
			}

			// Get current working directory for project name
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			projectName := filepath.Base(cwd)

			// Create notes manager
			notesManager, err := notes.NewNotesManager()
			if err != nil {
				return fmt.Errorf("failed to create notes manager: %w", err)
			}

			// Create new note
			note := &notes.Note{
				BaseRecord: notes.BaseRecord{
					Timestamp: time.Now(),
				},
				Content: content,
				Metadata: map[string]interface{}{
					"project": projectName,
					"type":    "remember",
				},
			}

			// Get current user
			username := os.Getenv("USER")
			if username == "" {
				username = "default"
			}

			// Save note
			if err := notesManager.SaveUserNote(username, note); err != nil {
				return fmt.Errorf("failed to save note: %w", err)
			}

			fmt.Printf("Saved successfully at %s\n", note.Timestamp.Format(time.RFC3339))
			return nil
		},
	}
}
