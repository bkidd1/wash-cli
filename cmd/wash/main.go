package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bkidd1/wash-cli/cmd/wash/bug"
	configcmd "github.com/bkidd1/wash-cli/cmd/wash/config"
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
- Monitor your project
- Provide intelligent suggestions
- Remember important context`,
}

func init() {
	// Add commands
	rootCmd.AddCommand(file.Command())
	rootCmd.AddCommand(bug.Command())
	rootCmd.AddCommand(monitor.Command())
	rootCmd.AddCommand(configcmd.Command())
	rootCmd.AddCommand(project.Command())
	rootCmd.AddCommand(
		remember.Command(),
		summary.Command(),
	)
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

	// Add pre-run function to check for API key
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip API key check for config and version commands
		if cmd.Use == "config" || cmd.Use == "version" {
			return nil
		}

		// Check if API key is set
		hasKey, err := config.ValidateAPIKey()
		if err != nil {
			return fmt.Errorf("failed to validate API key: %w", err)
		}

		if !hasKey {
			fmt.Println("\nWelcome to Wash CLI! 🚀")
			fmt.Println("Before you can start using Wash, you need to set up your OpenAI API key.")
			fmt.Println("You can get your API key from: https://platform.openai.com/api-keys")
			fmt.Println("\nWould you like to set up your API key now? (yes/no)")

			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			input = strings.TrimSpace(strings.ToLower(input))
			if input != "yes" && input != "y" {
				fmt.Println("\nYou'll need to set up your API key before using Wash.")
				fmt.Println("You can do this later by running: wash config set-key")
				os.Exit(0)
			}

			// Prompt for API key
			fmt.Print("Enter your OpenAI API key: ")
			apiKey, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			apiKey = strings.TrimSpace(apiKey)

			// Validate API key
			if apiKey == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			// Load current config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Update API key
			cfg.OpenAIKey = apiKey

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println("\nGreat! Your API key has been set up successfully.")
			fmt.Println("\nQuick Start Guide:")
			fmt.Println("-----------------")
			fmt.Println("Here are some common commands to get started:")
			fmt.Println("")
			fmt.Println("1. Analyze a file:")
			fmt.Println("   wash file <path/to/file>")
			fmt.Println("   Example: wash file main.go")
			fmt.Println("")
			fmt.Println("2. Monitor your development workflow:")
			fmt.Println("   wash monitor start")
			fmt.Println("   This will track your coding activities and provide insights")
			fmt.Println("")
			fmt.Println("3. Get a summary of your progress:")
			fmt.Println("   wash summary")
			fmt.Println("   Shows a summary of your recent development activities")
			fmt.Println("")
			fmt.Println("4. Remember important information:")
			fmt.Println("   wash remember \"Your note here\"")
			fmt.Println("   Stores important information for later reference")
			fmt.Println("")
			fmt.Println("5. Analyze a bug or issue:")
			fmt.Println("   wash bug \"Description of the bug\"")
			fmt.Println("   Provides analysis and potential solutions")
			fmt.Println("")
			fmt.Println("6. Analyze project structure:")
			fmt.Println("   wash project")
			fmt.Println("   Analyzes your project's organization and architecture")
			fmt.Println("")
			fmt.Println("For more information about any command, use:")
			fmt.Println("wash [command] --help")
			fmt.Println("")
			fmt.Println("Try running one of these commands to get started!")
			os.Exit(0)
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
