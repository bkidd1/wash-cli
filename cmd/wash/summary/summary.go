package summary

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/brinleekidd/wash-cli/internal/analyzer"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/spf13/cobra"
)

// Command creates the chat summary command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Analyze chat history and provide insights",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create analyzer
			analyzer := analyzer.NewAnalyzer(cfg.OpenAIKey)

			// Get the current working directory to create project-specific path
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			// Read the chat analysis file from project-specific directory
			projectPath := filepath.Base(cwd)
			notesDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", projectPath, "notes")
			analysisPath := filepath.Join(notesDir, "chat_analysis.txt")

			content, err := os.ReadFile(analysisPath)
			if err != nil {
				return fmt.Errorf("failed to read chat analysis file: %w", err)
			}

			// Get a summary of the analysis
			result, err := analyzer.AnalyzeChatSummary(context.Background(), string(content))
			if err != nil {
				return fmt.Errorf("failed to analyze chat summary: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	return cmd
}
