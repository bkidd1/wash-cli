package main

import (
	"fmt"
	"os"

	"github.com/bkidd1/wash-cli/cmd/wash/bug"
	"github.com/bkidd1/wash-cli/cmd/wash/file"
	"github.com/bkidd1/wash-cli/cmd/wash/monitor"
	"github.com/bkidd1/wash-cli/cmd/wash/project"
	"github.com/bkidd1/wash-cli/cmd/wash/remember"
	"github.com/bkidd1/wash-cli/cmd/wash/summary"
	versioncmd "github.com/bkidd1/wash-cli/cmd/wash/version"
	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/spf13/cobra"
)

//go:generate go build -o ../../wash

var rootCmd = &cobra.Command{
	Use:   "wash",
	Short: "Wash CLI - Your AI-powered development assistant",
	Long: `Wash CLI is an AI-powered development assistant that helps you:
- Analyze code and bugs
- Provide intelligent suggestions
- Remember important context`,
}

func init() {
	// Add commands
	rootCmd.AddCommand(file.Command())
	rootCmd.AddCommand(bug.Command())
	rootCmd.AddCommand(versioncmd.Command())

	// Add hidden commands
	monitorCmd := monitor.Command()
	monitorCmd.Hidden = true
	rootCmd.AddCommand(monitorCmd)

	rootCmd.AddCommand(project.Command())

	// Add hidden commands
	rememberCmd := remember.Command()
	summaryCmd := summary.Command()
	summaryCmd.Hidden = true
	rootCmd.AddCommand(rememberCmd, summaryCmd)

	// Hide the default completion command
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Set a custom help template
	rootCmd.SetHelpTemplate(`{{with .Long}}{{. | trimTrailingWhitespaces}}

{{end}}Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`)

	// Set a custom usage template
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	// Add pre-run function to check for API key
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip API key check for config and version commands
		if cmd.Use == "config" || cmd.Use == "version" {
			return nil
		}

		// Check if API key is set
		hasKey, err := config.ValidateAPIKey()
		if err != nil {
			return fmt.Errorf("error checking API key: %w", err)
		}

		if !hasKey {
			fmt.Println("No OpenAI API key found. Please set your API key to continue.")
			fmt.Println("You can do this later by running: wash config set-key")
			return fmt.Errorf("API key not set")
		}

		return nil
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
