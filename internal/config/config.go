package config

import (
	"fmt"
	"os"

	"governance-alerts-cosmos/internal/types"

	"github.com/spf13/viper"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*types.Config, error) {
	// Set default config file if not provided
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Set config file
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Read environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Create config struct
	var config types.Config

	// Unmarshal config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validateConfig validates the configuration
func validateConfig(config *types.Config) error {
	// Validate alert settings
	if config.Alerts.HoursBeforeStart <= 0 {
		return fmt.Errorf("hours_before_start must be greater than 0")
	}
	if config.Alerts.HoursBeforeEnd <= 0 {
		return fmt.Errorf("hours_before_end must be greater than 0")
	}
	if config.Alerts.CheckIntervalMinutes <= 0 {
		return fmt.Errorf("check_interval_minutes must be greater than 0")
	}

	// Validate networks
	if len(config.Networks) == 0 {
		return fmt.Errorf("at least one network must be configured")
	}

	for name, network := range config.Networks {
		if network.Name == "" {
			return fmt.Errorf("network name is required for %s", name)
		}
		if network.RestEndpoint == "" {
			return fmt.Errorf("rest_endpoint is required for network %s", name)
		}
		if network.ChainID == "" {
			return fmt.Errorf("chain_id is required for network %s", name)
		}
	}

	return nil
}
