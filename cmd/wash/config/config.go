package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/spf13/cobra"
)

// Command returns the config command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Wash CLI configuration",
		Long:  `Manage Wash CLI configuration settings including API keys and project settings.`,
	}

	// Add subcommands
	cmd.AddCommand(setKeyCommand())
	cmd.AddCommand(showConfigCommand())

	return cmd
}

// setKeyCommand returns the command to set/reset the API key
func setKeyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-key",
		Short: "Set or reset your OpenAI API key",
		Long:  `Set or reset your OpenAI API key. This will update the key in your configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load current config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get new API key from user
			fmt.Print("Enter your OpenAI API key: ")
			reader := bufio.NewReader(os.Stdin)
			apiKey, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			apiKey = strings.TrimSpace(apiKey)

			// Validate API key is not empty
			if apiKey == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			// Update config with new key
			cfg.OpenAIKey = apiKey
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println("API key updated successfully!")
			return nil
		},
	}
}

// showConfigCommand returns the command to show current configuration
func showConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Show the current configuration settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Print configuration
			fmt.Println("Current Configuration:")
			fmt.Println("---------------------")
			fmt.Printf("OpenAI API Key: %s\n", maskAPIKey(cfg.OpenAIKey))
			fmt.Printf("Project Goal: %s\n", cfg.ProjectGoal)
			fmt.Printf("Remember Notes: %d notes\n", len(cfg.RememberNotes))

			return nil
		},
	}
}

// maskAPIKey masks the API key for display
func maskAPIKey(key string) string {
	if key == "" {
		return "Not set"
	}
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
