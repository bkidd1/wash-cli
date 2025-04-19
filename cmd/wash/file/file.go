package file

import (
	"context"
	"fmt"

	"github.com/brinleekidd/wash-cli/internal/analyzer"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/spf13/cobra"
)

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

			// Create analyzer
			analyzer := analyzer.NewAnalyzer(cfg.OpenAIKey)

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

	return cmd
}
