package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Wash CLI configuration",
		Long: `Manage Wash CLI configuration settings including API keys and preferences.
This command allows you to:
- Set your OpenAI API key
- View current configuration
- Reset configuration to defaults`,
	}

	// Add subcommands
	cmd.AddCommand(setKeyCmd())
	cmd.AddCommand(showCmd())
	cmd.AddCommand(resetCmd())

	return cmd
}

func setKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-key",
		Short: "Set your OpenAI API key",
		Long: `Set your OpenAI API key for Wash CLI.
The key will be stored securely in your home directory and never committed to git.
You can also set the key using the OPENAI_API_KEY environment variable.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var apiKey string
			if len(args) > 0 {
				apiKey = args[0]
			} else {
				// Interactive mode
				fmt.Print("Enter your OpenAI API key: ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				apiKey = strings.TrimSpace(input)
			}

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

			fmt.Println("API key saved successfully!")
			return nil
		},
	}

	return cmd
}

func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Show the current Wash CLI configuration settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Print config
			fmt.Println("Current Configuration:")
			fmt.Println("---------------------")
			fmt.Printf("OpenAI API Key: %s\n", maskKey(cfg.OpenAIKey))
			fmt.Printf("Project Goal: %s\n", cfg.ProjectGoal)
			fmt.Printf("Remember Notes: %d\n", len(cfg.RememberNotes))
			return nil
		},
	}

	return cmd
}

func resetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long:  `Reset all Wash CLI configuration settings to their default values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create default config
			cfg := &config.Config{
				OpenAIKey:     "",
				ProjectGoal:   "",
				RememberNotes: []string{},
			}

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println("Configuration reset to defaults")
			return nil
		},
	}

	return cmd
}

// maskKey masks all but the last 4 characters of an API key
func maskKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(key)-4) + key[len(key)-4:]
}
