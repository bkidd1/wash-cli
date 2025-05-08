package file

import (
	"bufio"
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
					// Try to get the currently open file from common editor environment variables
					for _, envVar := range []string{"VSCODE_PID", "VSCODE_CWD", "VSCODE_IPC_HOOK", "VSCODE_NLS_CONFIG"} {
						if os.Getenv(envVar) != "" {
							// VS Code is running, try to get the active file
							if activeFile := os.Getenv("VSCODE_ACTIVE_FILE"); activeFile != "" {
								path = activeFile
								break
							}
						}
					}

					// If still no file found, try to get the active file from other editors
					if path == "." {
						for _, envVar := range []string{"EDITOR", "VISUAL"} {
							if editor := os.Getenv(envVar); editor != "" {
								// Check if it's a common editor that supports getting active file
								if strings.Contains(strings.ToLower(editor), "code") ||
									strings.Contains(strings.ToLower(editor), "vscode") {
									if activeFile := os.Getenv("VSCODE_ACTIVE_FILE"); activeFile != "" {
										path = activeFile
										break
									}
								}
							}
						}
					}

					// If no file was found, prompt the user for a file path
					if path == "." {
						fmt.Print("Enter file path: ")
						reader := bufio.NewReader(os.Stdin)
						input, err := reader.ReadString('\n')
						if err != nil {
							return fmt.Errorf("failed to read input: %w", err)
						}
						path = strings.TrimSpace(input)
						if path == "" {
							return fmt.Errorf("file path cannot be empty")
						}
					}
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

			// Check if this is a partial analysis
			if strings.Contains(result, "Would you like to analyze the remaining lines?") {
				fmt.Print("\nYour choice (y/n): ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}

				input = strings.TrimSpace(strings.ToLower(input))
				if input == "y" || input == "yes" {
					// Create a new channel for the second analysis
					done = make(chan bool)
					go loadingAnimation(done)

					// Get the remaining content
					content, err := os.ReadFile(absPath)
					if err != nil {
						done <- true
						return fmt.Errorf("error reading file: %w", err)
					}

					lines := strings.Split(string(content), "\n")
					// Assuming average of 6 tokens per line and reserving 4000 tokens for system prompt and overhead
					approxLines := (8192 - 4000) / 6 // GPT-4's context window is 8192 tokens

					// Further reduce by 30% to be safe
					approxLines = (approxLines * 7) / 10

					// Ensure we don't exceed the number of lines
					if approxLines > len(lines) {
						approxLines = len(lines)
					}

					remainingContent := strings.Join(lines[approxLines:], "\n")

					// Analyze the remaining content
					remainingResult, err := analyzer.AnalyzeContent(context.Background(), remainingContent)
					if err != nil {
						done <- true
						return fmt.Errorf("failed to analyze remaining content: %w", err)
					}

					done <- true
					fmt.Println("\nRemaining Analysis:")
					fmt.Println("------------------")
					fmt.Println(remainingResult)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&goal, "goal", "", "Specific goal for the file analysis")

	return cmd
}
