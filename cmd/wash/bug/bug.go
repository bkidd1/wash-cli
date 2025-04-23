package bug

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
	projectName string
	priority    string
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
			fmt.Printf("\rWashing bug... %s", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Command creates the bug command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bug [description]",
		Short: "Report and analyze a bug in your code",
		Long: `Report a bug in your code and get AI-assisted analysis and potential solutions.

This command will:
1. Prompt you for a description of the bug
2. Analyze your project context and recent changes
3. Suggest potential causes and solutions
4. Save the bug report with analysis for future reference

The analysis includes:
- Root cause analysis
- Impact assessment
- Suggested fixes
- Prevention strategies
- Related context

Examples:
  # Report a bug interactively
  wash bug

  # Report a bug directly
  wash bug "API endpoint returns 500 error"

  # Report a bug with priority
  wash bug --priority high "Critical security vulnerability"

  # Report a bug for specific project
  wash bug --project my-project "Database connection issues"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var description string
			if len(args) > 0 {
				description = strings.TrimSpace(strings.Join(args, " "))
			} else {
				// Get bug description from user
				fmt.Print("Please describe the bug you're experiencing: ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				description = strings.TrimSpace(input)
			}

			// Validate description
			if description == "" {
				return fmt.Errorf("bug description cannot be empty")
			}

			// Get project name
			if projectName == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
				projectName = filepath.Base(cwd)
			}

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create analyzer with project context
			analyzer := analyzer.NewTerminalAnalyzer(cfg.OpenAIKey, cfg.ProjectGoal, cfg.RememberNotes)

			// Create a channel to signal when analysis is done
			done := make(chan bool)
			go loadingAnimation(done)

			// Analyze the bug
			analysis, err := analyzer.AnalyzeBug(context.Background(), description)
			if err != nil {
				done <- true
				return fmt.Errorf("failed to analyze bug: %w", err)
			}

			// Signal that analysis is complete
			done <- true

			// Create project-specific bug directory
			bugDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", projectName, "bugs")
			if err := os.MkdirAll(bugDir, 0755); err != nil {
				return fmt.Errorf("failed to create bugs directory: %w", err)
			}

			// Generate bug report filename with timestamp
			timestamp := time.Now().Format("2006-01-02-15-04-05")
			bugFile := filepath.Join(bugDir, fmt.Sprintf("bug_%s.md", timestamp))

			// Create bug report with analysis
			report := fmt.Sprintf(`# Bug Report
*Reported on %s*

## Description
%s

## Analysis
%s

## Potential Causes
%s

## Suggested Solutions
%s

## Related Context
%s

## Priority
%s

## Status
Open

## Notes
`,
				time.Now().Format("2006-01-02 15:04:05"),
				description,
				analysis.Analysis,
				analysis.PotentialCauses,
				analysis.SuggestedSolutions,
				analysis.RelatedContext,
				priority,
			)

			// Save bug report
			if err := os.WriteFile(bugFile, []byte(report), 0644); err != nil {
				return fmt.Errorf("failed to save bug report: %w", err)
			}

			// Print analysis to console
			fmt.Println("\nBug Analysis Results:")
			fmt.Println("-------------------")
			fmt.Printf("\nAnalysis:\n%s\n", analysis.Analysis)
			fmt.Printf("\nPotential Causes:\n%s\n", analysis.PotentialCauses)
			fmt.Printf("\nSuggested Solutions:\n%s\n", analysis.SuggestedSolutions)
			fmt.Printf("\nBug report saved to: %s\n", bugFile)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "Project name (defaults to current directory name)")
	cmd.Flags().StringVar(&priority, "priority", "medium", "Bug priority (low, medium, high)")

	return cmd
}
