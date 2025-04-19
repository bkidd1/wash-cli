package project

import (
	"context"
	"fmt"

	"github.com/brinleekidd/wash-cli/internal/analyzer"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/spf13/cobra"
)

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

			// Analyze project structure
			result, err := analyzer.AnalyzeProjectStructure(context.Background(), dirPath)
			if err != nil {
				return fmt.Errorf("failed to analyze project structure: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	return cmd
}
