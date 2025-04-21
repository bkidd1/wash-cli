package note

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bkidd1/wash-cli/internal/notes"
	"github.com/spf13/cobra"
)

// Command returns the note command
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "note [content]",
		Short: "Create a note",
		Long:  `Create a note for your project. Notes are stored in ~/.wash/projects/[project-name]/notes/`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current working directory for project name
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %v", err)
			}
			projectName := filepath.Base(cwd)

			// Create notes manager
			notesManager, err := notes.NewNotesManager()
			if err != nil {
				return fmt.Errorf("failed to create notes manager: %v", err)
			}

			// Create new note
			note := &notes.Note{
				Type:      "user",
				Content:   args[0],
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"project": projectName,
				},
			}

			// Get current user
			username := os.Getenv("USER")
			if username == "" {
				username = "default"
			}

			// Save note
			if err := notesManager.SaveUserNote(username, note); err != nil {
				return fmt.Errorf("failed to save note: %v", err)
			}

			fmt.Printf("Note created successfully at %s\n", note.Timestamp.Format(time.RFC3339))
			return nil
		},
	}
}
