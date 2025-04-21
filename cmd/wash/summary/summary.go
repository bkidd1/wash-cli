package summary

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bkidd1/wash-cli/internal/notes"
	"github.com/spf13/cobra"
)

// Command returns the summary command
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Get a summary of your recent progress",
		Long: `Get a summary of your recent progress, including recent notes, decisions, and changes.
This command provides an overview of your project's recent activity.`,
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

			// Get recent interactions
			interactions, err := notesManager.LoadInteractions(projectName)
			if err != nil {
				return fmt.Errorf("failed to load interactions: %v", err)
			}

			// Print summary
			fmt.Printf("Recent Activity Summary for %s:\n", projectName)
			fmt.Println("----------------------------------------")

			if len(interactions) == 0 {
				fmt.Println("No recent activity found.")
				return nil
			}

			// Sort interactions by timestamp (newest first)
			for i := len(interactions) - 1; i >= 0; i-- {
				interaction := interactions[i]
				if time.Since(interaction.Timestamp) <= 24*time.Hour {
					fmt.Printf("[%s] %s: %s\n",
						interaction.Timestamp.Format(time.RFC3339),
						interaction.Type,
						interaction.Content.UserInput)
				}
			}

			return nil
		},
	}
}
