package bug

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	description      string
	stepsToReproduce []string
	expectedBehavior string
	actualBehavior   string
	priority         string
)

// Command creates the bug command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bug",
		Short: "Report a bug in your code",
		Long:  `Report a bug in your code with detailed information about the issue.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current working directory for project context
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %v", err)
			}

			// Create project-specific bug directory
			projectPath := filepath.Base(cwd)
			bugDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", projectPath, "bugs")
			if err := os.MkdirAll(bugDir, 0755); err != nil {
				return fmt.Errorf("failed to create bugs directory: %v", err)
			}

			// Generate bug report filename with timestamp
			timestamp := time.Now().Format("2006-01-02-15-04-05")
			bugFile := filepath.Join(bugDir, fmt.Sprintf("bug_%s.md", timestamp))

			// Create bug report
			report := fmt.Sprintf(`# Bug Report
*Reported on %s*

## Description
%s

## Steps to Reproduce
%s

## Expected Behavior
%s

## Actual Behavior
%s

## Priority
%s

## Status
Open

## Notes
`,
				time.Now().Format("2006-01-02 15:04:05"),
				description,
				formatSteps(stepsToReproduce),
				expectedBehavior,
				actualBehavior,
				priority,
			)

			// Save bug report
			if err := os.WriteFile(bugFile, []byte(report), 0644); err != nil {
				return fmt.Errorf("failed to save bug report: %v", err)
			}

			fmt.Printf("Bug report saved to %s\n", bugFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the bug")
	cmd.Flags().StringSliceVarP(&stepsToReproduce, "steps", "s", []string{}, "Steps to reproduce the bug")
	cmd.Flags().StringVarP(&expectedBehavior, "expected", "e", "", "Expected behavior")
	cmd.Flags().StringVarP(&actualBehavior, "actual", "a", "", "Actual behavior")
	cmd.Flags().StringVarP(&priority, "priority", "p", "medium", "Priority level (low, medium, high)")

	// Mark required flags
	cmd.MarkFlagRequired("description")
	cmd.MarkFlagRequired("expected")
	cmd.MarkFlagRequired("actual")

	return cmd
}

func formatSteps(steps []string) string {
	if len(steps) == 0 {
		return "No steps provided"
	}
	var result string
	for i, step := range steps {
		result += fmt.Sprintf("%d. %s\n", i+1, step)
	}
	return result
}
