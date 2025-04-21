package main

import (
	"fmt"
	"os"

	"github.com/bkidd1/wash-cli/cmd/wash/bug"
	"github.com/bkidd1/wash-cli/cmd/wash/file"
	"github.com/bkidd1/wash-cli/cmd/wash/monitor"
	"github.com/bkidd1/wash-cli/cmd/wash/project"
	"github.com/bkidd1/wash-cli/cmd/wash/remember"
	"github.com/bkidd1/wash-cli/cmd/wash/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wash",
	Short: "Wash CLI - A development assistant",
	Long:  `Wash CLI is a development assistant that helps track errors, decisions, and project state.`,
}

func initCommands() error {
	// Add commands
	rootCmd.AddCommand(file.Command())
	rootCmd.AddCommand(project.Command())
	rootCmd.AddCommand(monitor.Command())
	rootCmd.AddCommand(bug.Command())
	rootCmd.AddCommand(version.Command())
	rootCmd.AddCommand(remember.Command())

	// Hide the default completion command as we don't need it
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	return nil
}

func main() {
	if err := initCommands(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing commands: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
