package config

import (
	"github.com/spf13/viper"
)

// LoadConfig loads the configuration using the modular config loader
// This is the main entry point for loading configuration
func LoadConfig(serviceName string) (*viper.Viper, error) {
	loader := NewConfigLoader()
	return loader.LoadConfig(serviceName)
} 