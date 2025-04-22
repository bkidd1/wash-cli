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
		Use:   "bug",
		Short: "Report and analyze a bug in your code",
		Long: `Report a bug in your code and get AI-assisted analysis and potential solutions.

This command will:
1. Prompt you for a description of the bug
2. Analyze your project context and recent changes
3. Suggest potential causes and solutions
4. Save the bug report with analysis for future reference`,
		Run: func(cmd *cobra.Command, args []string) {
			// Get bug description from user
			fmt.Print("Please describe the bug you're experiencing: ")
			reader := bufio.NewReader(os.Stdin)
			description, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
				os.Exit(1)
			}
			description = strings.TrimSpace(description)

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
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
				fmt.Fprintf(os.Stderr, "Error analyzing bug: %v\n", err)
				os.Exit(1)
			}

			// Signal that analysis is complete
			done <- true

			// Get current working directory for project context
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
				os.Exit(1)
			}

			// Create project-specific bug directory
			projectPath := filepath.Base(cwd)
			bugDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", projectPath, "bugs")
			if err := os.MkdirAll(bugDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating bugs directory: %v\n", err)
				os.Exit(1)
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
			)

			// Save bug report
			if err := os.WriteFile(bugFile, []byte(report), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving bug report: %v\n", err)
				os.Exit(1)
			}

			// Print analysis to console
			fmt.Printf("\nAnalysis:\n%s\n", analysis.Analysis)
			fmt.Printf("\nPotential Causes:\n%s\n", analysis.PotentialCauses)
			fmt.Printf("\nSuggested Solutions:\n%s\n", analysis.SuggestedSolutions)
		},
	}

	return cmd
}
