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
	OpenAIKey     string
	LogPath       string
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
			viper.Set("log_path", filepath.Join(configDir, "logs"))
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

	if openAIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not found. Please set OPENAI_API_KEY environment variable or add it to ~/.wash/wash.yaml")
	}

	// Get log path from config file or use default
	logPath := viper.GetString("log_path")
	if logPath == "" {
		logPath = filepath.Join(configDir, "logs")
	}

	// Get project goal and remember notes
	projectGoal := viper.GetString("project_goal")
	rememberNotes := viper.GetStringSlice("remember_notes")

	return &Config{
		OpenAIKey:     openAIKey,
		LogPath:       logPath,
		ProjectGoal:   projectGoal,
		RememberNotes: rememberNotes,
	}, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	viper.Set("openai_key", config.OpenAIKey)
	viper.Set("log_path", config.LogPath)
	viper.Set("project_goal", config.ProjectGoal)
	viper.Set("remember_notes", config.RememberNotes)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, DefaultConfigName+"."+DefaultConfigType)
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
