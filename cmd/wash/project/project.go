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
			fmt.Printf("\rAnalyzing project... %s", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Command creates the project structure analysis command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage project settings and analysis",
		Long: `Manage project settings and analyze project structure for optimization opportunities.
Provides insights into potential refactoring opportunities and project organization.`,
	}

	// Add subcommands
	cmd.AddCommand(analyzeCommand())
	cmd.AddCommand(goalCommand())

	return cmd
}

// analyzeCommand creates the project analysis command
func analyzeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [path]",
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
			analyzer := analyzer.NewAnalyzer(cfg.OpenAIKey, cfg.ProjectGoal, cfg.RememberNotes)

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
}

// goalCommand creates the project goal command
func goalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "goal [goal]",
		Short: "Set or view the project goal",
		Long: `Set or view the project goal. This goal will be used to provide context
for code analysis and suggestions.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if len(args) == 0 {
				// View current goal
				if cfg.ProjectGoal == "" {
					fmt.Println("No project goal set. Use 'wash project goal <goal>' to set one.")
				} else {
					fmt.Printf("Current project goal: %s\n", cfg.ProjectGoal)
				}
				return nil
			}

			// Set new goal
			cfg.ProjectGoal = args[0]
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Project goal set to: %s\n", cfg.ProjectGoal)
			return nil
		},
	}
}
