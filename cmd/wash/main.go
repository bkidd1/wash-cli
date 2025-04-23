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
	"github.com/bkidd1/wash-cli/pkg/version"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "wash",
	Short: "Wash CLI - Your AI-powered development assistant",
	Long: `Wash CLI is an AI-powered development assistant that helps you:
- Monitor your development workflow
- Track project progress
- Remember important details
- Analyze code and suggest improvements
- Generate summaries and documentation
- Debug and fix issues

For more information about a specific command, use 'wash [command] --help'`,
	Version: version.String(),
}

func init() {
	// Add commands
	rootCmd.AddCommand(project.Command())
	rootCmd.AddCommand(file.Command())
	rootCmd.AddCommand(monitor.Command())
	rootCmd.AddCommand(remember.Command())
	rootCmd.AddCommand(summary.Command())
	rootCmd.AddCommand(bug.Command())
	rootCmd.AddCommand(versioncmd.Command())

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
}
