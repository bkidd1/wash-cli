package file

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/analyzer"
	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/spf13/cobra"
)

// loadingAnimation shows a simple loading animation
func loadingAnimation(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Printf("\r") // Clear the line
			return
		default:
			fmt.Printf("\rWashing file... %s", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Command creates the file analysis command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file [path]",
		Short: "Wash a single file for optimization opportunities",
		Long: `Washes the specified file and suggests alternative coding pathways
and optimizations. The washing focuses on code structure, performance,
and maintainability. If no file path is provided, the currently open file will be used.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the path to analyze
			path := "."
			if len(args) > 0 {
				path = args[0]
			} else {
				// Try to get the currently open file from the editor
				if selectedFile := os.Getenv("WASH_SELECTED_FILE"); selectedFile != "" {
					path = selectedFile
				} else {
					// If no file is selected, try to get the current file from the editor
					// This is a placeholder - we'll need to implement the actual editor integration
					fmt.Println("No file path provided and no file is currently open in the editor.")
					fmt.Println("Please either provide a file path or open a file in your editor.")
					return nil
				}
			}

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create analyzer with project context
			analyzer := analyzer.NewTerminalAnalyzer(cfg.OpenAIKey, cfg.ProjectGoal, cfg.RememberNotes)

			// Create a channel to signal when washing is done
			done := make(chan bool)
			go loadingAnimation(done)

			// Wash file
			result, err := analyzer.AnalyzeFile(context.Background(), path)
			if err != nil {
				done <- true
				return fmt.Errorf("failed to wash file: %w", err)
			}

			// Signal that washing is complete
			done <- true

			// Print results
			fmt.Println(result)
			return nil
		},
	}

	return cmd
}
