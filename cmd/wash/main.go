package main

import (
	"fmt"
	"os"

	"github.com/bkidd1/wash-cli/cmd/wash/file"
	"github.com/bkidd1/wash-cli/cmd/wash/monitor"
	"github.com/bkidd1/wash-cli/cmd/wash/project"
	"github.com/bkidd1/wash-cli/cmd/wash/remember"
	"github.com/bkidd1/wash-cli/cmd/wash/summary"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wash",
	Short: "Wash CLI - A development assistant",
	Long:  `Wash Your Code!`,
}

func initCommands() error {
	// Add commands
	rootCmd.AddCommand(project.Command())
	rootCmd.AddCommand(file.Command())
	rootCmd.AddCommand(monitor.Command())
	rootCmd.AddCommand(remember.Command())
	rootCmd.AddCommand(summary.Command())

	// Hide the default completion command as we don't need it
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Hide the flags section
	rootCmd.SetHelpTemplate(`{{.Long}}

Call any of the following commands with "wash [command]"
{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`)

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
