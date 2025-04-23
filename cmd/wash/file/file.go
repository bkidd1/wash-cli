package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
		Short: "Analyze and optimize a single file",
		Long: `Analyzes the specified file and suggests improvements for:
- Code structure
- Performance
- Maintainability
- Best practices
- Security
- Error handling

The analysis provides:
1. Code quality assessment
2. Optimization suggestions
3. Alternative implementations
4. Best practice recommendations

Examples:
  # Analyze current file in editor
  wash file

  # Analyze specific file
  wash file main.go

  # Analyze with specific goal
  wash file --goal "Improve error handling and logging" main.go`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the path to analyze
			path := "."
			if len(args) > 0 {
				path = args[0]
			} else {
				// Try to get the currently open file from the editor
				if selectedFile := os.Getenv("WASH_SELECTED_FILE"); selectedFile != "" {
					path = selectedFile
				} else {
					return fmt.Errorf("no file path provided and no file is currently open in the editor")
				}
			}

			// Validate path exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", path)
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
			analyzer := analyzer.NewTerminalAnalyzer(cfg.OpenAIKey, cfg.ProjectGoal, cfg.RememberNotes)

			// Create a channel to signal when analysis is done
			done := make(chan bool)
			go loadingAnimation(done)

			// Analyze file
			result, err := analyzer.AnalyzeFile(context.Background(), absPath)
			if err != nil {
				done <- true
				return fmt.Errorf("failed to analyze file: %w", err)
			}

			// Signal that analysis is complete
			done <- true

			// Print results
			fmt.Println("\nAnalysis Results:")
			fmt.Println("----------------")
			fmt.Println(result)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&goal, "goal", "", "Specific goal for the file analysis")

	return cmd
}
