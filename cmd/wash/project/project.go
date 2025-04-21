package project

import (
	"context"
	"fmt"
	"time"

	"github.com/brinleekidd/wash-cli/internal/analyzer"
	"github.com/brinleekidd/wash-cli/pkg/config"
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
			fmt.Printf("\rAnalyzing project... %s", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Command creates the project structure analysis command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [path]",
		Short: "Analyze project structure for optimization opportunities",
		Long: `Analyzes the project structure and suggests improvements for
organization, modularity, and maintainability. Provides insights into
potential refactoring opportunities.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create analyzer
			analyzer := analyzer.NewAnalyzer(cfg.OpenAIKey)

			// Get directory path
			dirPath := "."
			if len(args) > 0 {
				dirPath = args[0]
			}

			// Create a channel to signal when analysis is done
			done := make(chan bool)
			go loadingAnimation(done)

			// Analyze project structure
			result, err := analyzer.AnalyzeProjectStructure(context.Background(), dirPath)
			if err != nil {
				done <- true
				return fmt.Errorf("failed to analyze project structure: %w", err)
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
