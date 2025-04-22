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
		Use:   "project",
		Short: "Manage project settings and analysis",
		Long:  `Manage project settings and analyze project structure for optimization opportunities.`,
	}

	cmd.AddCommand(analyzeCommand())
	cmd.AddCommand(goalCommand())

	return cmd
}

// analyzeCommand creates the analyze subcommand
func analyzeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [path]",
		Short: "Analyze project structure for optimization opportunities",
		Long: `Analyzes the project structure and suggests improvements to organization,
architecture, and code structure. The analysis focuses on maintainability,
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

			// Analyze project structure
			result, err := analyzer.AnalyzeProjectStructure(context.Background(), path)
			if err != nil {
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			// Print results
			fmt.Println(result)
			return nil
		},
	}
}

// goalCommand creates the goal subcommand
func goalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "goal [goal]",
		Short: "Set or view the project goal",
		Long: `Sets or views the project goal. The goal is used to provide context
for analysis and suggestions. If no goal is provided, the current goal is displayed.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// If no goal provided, display current goal
			if len(args) == 0 {
				if cfg.ProjectGoal == "" {
					fmt.Println("No project goal set.")
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
