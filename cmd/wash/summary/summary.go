package summary

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/notes"
	"github.com/spf13/cobra"
)

var (
	// Flags
	projectName string
	date        string
)

// Command returns the summary command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Get a summary of project progress",
		Long: `Get a comprehensive summary of your project's progress, including:
- Recent changes and updates
- Key decisions and their impact
- File modifications
- Risk assessments
- Alternative approaches considered

The summary provides insights into:
1. What has been accomplished
2. Potential issues to watch for
3. Areas for improvement
4. Next steps

Examples:
  # Get summary for current project
  wash summary

  # Get summary for specific project
  wash summary --project my-project

  # Get summary for specific date
  wash summary --date 2024-04-23`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Get all progress notes
			progressNotes, err := notesManager.LoadProjectProgress(projectName)
			if err != nil {
				return fmt.Errorf("failed to load progress notes: %w", err)
			}

			if len(progressNotes) == 0 {
				fmt.Println("No progress notes found.")
				return nil
			}

			// Parse target date if provided
			var targetDate time.Time
			if date != "" {
				parsedDate, err := time.Parse("2006-01-02", date)
				if err != nil {
					return fmt.Errorf("invalid date format. Please use YYYY-MM-DD: %w", err)
				}
				targetDate = parsedDate
			} else {
				targetDate = time.Now().Truncate(24 * time.Hour)
			}

			// Find notes for target date
			var targetNotes []*notes.ProjectProgressNote
			for _, note := range progressNotes {
				if note.Timestamp.Truncate(24 * time.Hour).Equal(targetDate) {
					targetNotes = append(targetNotes, note)
				}
			}

			// If no notes for target date, find the most recent day with notes
			if len(targetNotes) == 0 {
				var mostRecent time.Time
				for _, note := range progressNotes {
					noteDate := note.Timestamp.Truncate(24 * time.Hour)
					if noteDate.After(mostRecent) {
						mostRecent = noteDate
						targetNotes = []*notes.ProjectProgressNote{note}
					} else if noteDate.Equal(mostRecent) {
						targetNotes = append(targetNotes, note)
					}
				}
				targetDate = mostRecent
				fmt.Printf("No notes found for specified date. Showing most recent notes from %s\n", targetDate.Format("2006-01-02"))
			}

			// Print summary
			fmt.Printf("\nProgress Summary for %s - %s\n", projectName, targetDate.Format("2006-01-02"))
			fmt.Println("----------------------------------------")

			// Section 1: Progress Made
			fmt.Println("\nProgress Made:")
			fmt.Println("-------------")
			for _, note := range targetNotes {
				fmt.Printf("- %s\n", note.Title)
				if note.Description != "" {
					fmt.Printf("  %s\n", note.Description)
				}
			}

			// Section 2: Potential Mistakes and Alternatives
			fmt.Println("\nPotential Mistakes and Alternatives:")
			fmt.Println("-----------------------------------")
			for _, note := range targetNotes {
				if note.Impact.RiskLevel != "" {
					fmt.Printf("In '%s':\n", note.Title)
					if note.Impact.Scope != "" {
						fmt.Printf("  Decision: %s\n", note.Impact.Scope)
					}
					fmt.Printf("  Risk Level: %s\n", note.Impact.RiskLevel)
					if len(note.Impact.AffectedAreas) > 0 {
						fmt.Println("  Alternative Approaches:")
						for _, area := range note.Impact.AffectedAreas {
							fmt.Printf("  - %s\n", area)
						}
					}
					fmt.Println()
				}
			}

			// Section 3: File Changes
			fmt.Println("\nFiles Changed:")
			fmt.Println("-------------")
			allFiles := make(map[string]bool)
			for _, note := range targetNotes {
				for _, file := range note.Changes.FilesModified {
					allFiles[file] = true
				}
			}
			if len(allFiles) == 0 {
				fmt.Println("No files were modified.")
			} else {
				for file := range allFiles {
					fmt.Printf("- %s\n", file)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "Project name (defaults to current directory name)")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date to show summary for (format: YYYY-MM-DD)")

	return cmd
}
