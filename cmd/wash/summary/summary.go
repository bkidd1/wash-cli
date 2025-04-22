package summary

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/notes"
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

			// Get all progress notes
			progressNotes, err := notesManager.LoadProjectProgress(projectName)
			if err != nil {
				return fmt.Errorf("failed to load progress notes: %v", err)
			}

			if len(progressNotes) == 0 {
				fmt.Println("No progress notes found.")
				return nil
			}

			// Find notes for today or most recent day
			today := time.Now().Truncate(24 * time.Hour)
			var targetNotes []*notes.ProjectProgressNote
			var targetDate time.Time

			// First try to find notes from today
			for _, note := range progressNotes {
				if note.Timestamp.Truncate(24 * time.Hour).Equal(today) {
					targetNotes = append(targetNotes, note)
				}
			}

			// If no notes from today, find the most recent day with notes
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
			} else {
				targetDate = today
			}

			// Print summary
			fmt.Printf("Progress Summary for %s - %s\n", projectName, targetDate.Format("2006-01-02"))
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
}
