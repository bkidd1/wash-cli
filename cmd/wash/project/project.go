package project

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
			fmt.Printf("\rWashing project... %s", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Command creates the project command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [path]",
		Short: "Wash project structure for optimization opportunities",
		Long: `Washes the project structure and suggests improvements to organization,
architecture, and code structure. The washing focuses on maintainability,
scalability, and best practices.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the path to analyze
			path := "."
			if len(args) > 0 {
				path = args[0]
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

			// Wash project structure
			result, err := analyzer.AnalyzeProjectStructure(context.Background(), path)
			if err != nil {
				done <- true
				return fmt.Errorf("failed to wash project: %w", err)
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
