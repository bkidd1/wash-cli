package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/analyzer"
	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/spf13/cobra"
)

var (
	// Flags
	goal string
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
		Short: "Analyze and optimize project structure",
		Long: `Analyzes your project structure and suggests improvements to organization,
architecture, and code structure. The analysis focuses on:

- Maintainability
- Scalability
- Code organization
- Best practices
- Performance optimization
- Security considerations

The command will:
1. Scan your project structure
2. Analyze code patterns and architecture
3. Identify potential improvements
4. Generate actionable recommendations

Examples:
  # Analyze current directory
  wash project

  # Analyze specific directory
  wash project ./src

  # Analyze with specific goal
  wash project --goal "Improve code organization and reduce technical debt"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the path to analyze
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			// Validate path exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", path)
			}

			// Get absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Override project goal if specified
			if goal != "" {
				cfg.ProjectGoal = goal
			}

			// Create analyzer with project context
			analyzer := analyzer.NewTerminalAnalyzer(cfg.OpenAIKey, cfg.ProjectGoal, nil)

			// Create a channel to signal when washing is done
			done := make(chan bool)
			go loadingAnimation(done)

			// Wash project structure
			result, err := analyzer.AnalyzeProjectStructure(context.Background(), absPath)
			if err != nil {
				// Check if error is token limit related
				if strings.Contains(err.Error(), "token") || strings.Contains(err.Error(), "length") {
					done <- true
					fmt.Println("\n⚠️  Project is too large for complete analysis.")
					fmt.Println("Please specify a subdirectory to analyze (e.g., 'cmd', 'internal', 'pkg'):")

					var subdir string
					fmt.Scanln(&subdir)

					// Validate the subdirectory exists
					subdirPath := filepath.Join(absPath, subdir)
					if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
						return fmt.Errorf("subdirectory does not exist: %s", subdir)
					}

					// Create a new channel for the subdirectory analysis
					done = make(chan bool)
					go loadingAnimation(done)

					// Analyze the subdirectory
					result, err = analyzer.AnalyzeProjectStructure(context.Background(), subdirPath)
					if err != nil {
						done <- true
						return fmt.Errorf("failed to analyze subdirectory: %w", err)
					}

					done <- true
					fmt.Printf("\nAnalysis Results for %s directory:\n", subdir)
					fmt.Println("-------------------------------")
					fmt.Println(result)
					return nil
				}

				done <- true
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			// Signal that washing is complete
			done <- true

			// Print results
			fmt.Println("\nAnalysis Results:")
			fmt.Println("----------------")
			fmt.Println(result)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&goal, "goal", "", "Specific goal for the project analysis")

	return cmd
}
