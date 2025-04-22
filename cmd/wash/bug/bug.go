package bug

import (
	"context"
	"fmt"
	"os"

	"github.com/bkidd1/wash-cli/internal/services/analyzer"
	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/spf13/cobra"
)

// Command creates the bug command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bug",
		Short: "Analyze and fix bugs in code",
		Long:  `Analyzes code for potential bugs and suggests fixes.`,
	}

	cmd.AddCommand(analyzeCommand())
	cmd.AddCommand(fixCommand())

	return cmd
}

// analyzeCommand creates the analyze subcommand
func analyzeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [file]",
		Short: "Analyze a file for potential bugs",
		Long: `Analyzes a file for potential bugs and suggests improvements.
The analysis focuses on common programming errors, edge cases, and best practices.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create analyzer with project context
			analyzer := analyzer.NewTerminalAnalyzer(cfg.OpenAIKey, cfg.ProjectGoal, cfg.RememberNotes)

			// Analyze file
			result, err := analyzer.AnalyzeFile(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("failed to analyze file: %w", err)
			}

			// Print results
			fmt.Println(result)
			return nil
		},
	}
}

// fixCommand creates the fix subcommand
func fixCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fix [file] [error]",
		Short: "Fix a specific error in a file",
		Long: `Fixes a specific error in a file by analyzing the error message
and suggesting a solution.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Read file content
			content, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			// Create analyzer with project context
			analyzer := analyzer.NewTerminalAnalyzer(cfg.OpenAIKey, cfg.ProjectGoal, cfg.RememberNotes)
			// Get error fix
			result, err := analyzer.GetErrorFix(context.Background(), string(content), args[1])
			if err != nil {
				return fmt.Errorf("failed to get error fix: %w", err)
			}

			// Print results
			fmt.Println(result)
			return nil
		},
	}
}
