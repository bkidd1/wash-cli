package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigName is the default name of the config file
	DefaultConfigName = ".wash"
	// DefaultConfigType is the default type of the config file
	DefaultConfigType = "yaml"
)

// Config holds the application configuration
type Config struct {
	OpenAIKey     string   `yaml:"openai_key"`
	ProjectGoal   string   `yaml:"project_goal,omitempty"`
	RememberNotes []string `yaml:"remember_notes,omitempty"`
}

// LoadConfig loads the configuration from file and environment variables
func LoadConfig() (*Config, error) {
	// Set up Viper
	viper.SetConfigName("wash")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.wash")

	// Create config directory if it doesn't exist
	configDir := filepath.Join(os.Getenv("HOME"), ".wash")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating config directory: %w", err)
	}

	// Try to read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create it with default values
			viper.Set("openai_key", "")
			viper.Set("project_goal", "")
			viper.Set("remember_notes", []string{})
			if err := viper.SafeWriteConfig(); err != nil {
				return nil, fmt.Errorf("error creating config file: %w", err)
			}
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Get OpenAI key from environment variable or config file
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		openAIKey = viper.GetString("openai_key")
	}

	// Get project goal and remember notes
	projectGoal := viper.GetString("project_goal")
	rememberNotes := viper.GetStringSlice("remember_notes")

	return &Config{
		OpenAIKey:     openAIKey,
		ProjectGoal:   projectGoal,
		RememberNotes: rememberNotes,
	}, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	// Reset Viper configuration
	viper.Reset()

	// Set up Viper again
	viper.SetConfigName("wash")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.wash")

	// Set the values
	viper.Set("openai_key", config.OpenAIKey)
	viper.Set("project_goal", config.ProjectGoal)
	viper.Set("remember_notes", config.RememberNotes)

	// Get the config file path
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".wash", "wash.yaml")

	// Write the config file
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateAPIKey checks if the API key is set and valid
func ValidateAPIKey() (bool, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return false, fmt.Errorf("failed to load config: %w", err)
	}

	// Check if API key is set
	if cfg.OpenAIKey == "" {
		return false, nil
	}

	// TODO: Add actual API key validation by making a test call to OpenAI
	// For now, we'll just check if it's not empty
	return true, nil
}
