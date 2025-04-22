package file

import (
	"context"
	"fmt"
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
		Short: "Analyze a single file for optimization opportunities",
		Long: `Analyzes the specified file and suggests alternative coding pathways
and optimizations. The analysis focuses on code structure, performance,
and maintainability.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create analyzer with project context
			analyzer := analyzer.NewTerminalAnalyzer(cfg.OpenAIKey, cfg.ProjectGoal, cfg.RememberNotes)

			// Create a channel to signal when analysis is done
			done := make(chan bool)
			go loadingAnimation(done)

			// Analyze file
			result, err := analyzer.AnalyzeFile(context.Background(), args[0])
			if err != nil {
				done <- true
				return fmt.Errorf("failed to analyze file: %w", err)
			}

			// Signal that analysis is complete
			done <- true

			// Print results
			fmt.Println(result)
			return nil
		},
	}

	return cmd
}
