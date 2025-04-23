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

var (
	// Flags
	projectName string
	tags        []string
)

// Command returns the remember command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remember [content]",
		Short: "Save important information to remember",
		Long: `Save important information, notes, or reminders for your project.
The remember command helps you keep track of:
- Important decisions
- Implementation details
- Future tasks
- Project-specific knowledge
- Development patterns

Notes are stored in ~/.wash/remember/[project-name]/

Examples:
  # Save a note interactively
  wash remember

  # Save a note directly
  wash remember "Implement caching for better performance"

  # Save a note with tags
  wash remember "Add error handling" --tags "error,security"

  # Save a note for specific project
  wash remember "Update documentation" --project my-project`,
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

			// Get project name
			if projectName == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
				projectName = filepath.Base(cwd)
			}

			// Create notes manager
			notesManager, err := notes.NewNotesManager()
			if err != nil {
				return fmt.Errorf("failed to create notes manager: %w", err)
			}

			// Create new note
			note := &notes.RememberNote{
				Timestamp: time.Now(),
				Content:   content,
				Metadata: map[string]interface{}{
					"project": projectName,
					"type":    "remember",
					"tags":    tags,
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

			fmt.Printf("\nNote saved successfully!\n")
			fmt.Printf("Time: %s\n", note.Timestamp.Format(time.RFC3339))
			fmt.Printf("Project: %s\n", projectName)
			if len(tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
			}
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "Project name (defaults to current directory name)")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags for the note (comma-separated)")

	return cmd
}
