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
	OpenAIAPIKey string `mapstructure:"openai_api_key"`
}

// LoadConfig loads the configuration from file and environment variables
func LoadConfig() (*Config, error) {
	// Set up Viper
	viper.SetConfigName(DefaultConfigName)
	viper.SetConfigType(DefaultConfigType)

	// Add config paths
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	viper.AddConfigPath(home)
	viper.AddConfigPath(".")

	// Read from environment variables
	viper.SetEnvPrefix("WASH")
	viper.AutomaticEnv()
	viper.BindEnv("openai_api_key")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required settings
	if config.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required. Set it in %s or WASH_OPENAI_API_KEY environment variable", filepath.Join(home, DefaultConfigName+"."+DefaultConfigType))
	}

	return &config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	viper.Set("openai_api_key", config.OpenAIAPIKey)

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
