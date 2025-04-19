package main

import (
	"fmt"
	"os"

	"github.com/brinleekidd/wash-cli/cmd/wash/bug"
	"github.com/brinleekidd/wash-cli/cmd/wash/chat"
	"github.com/brinleekidd/wash-cli/cmd/wash/file"
	"github.com/brinleekidd/wash-cli/cmd/wash/project"
	"github.com/brinleekidd/wash-cli/cmd/wash/summary"
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
	rootCmd.AddCommand(chat.Command())
	rootCmd.AddCommand(summary.Command())
	rootCmd.AddCommand(bug.Command())

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
