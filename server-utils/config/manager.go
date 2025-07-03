package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ConfigManager provides a simplified interface for loading, validating, and logging configuration
type ConfigManager struct {
	loader    ConfigLoader
	validator *ConfigValidator
	logger    *ConfigLogger
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		loader:    NewConfigLoader(),
		validator: NewConfigValidator(),
		logger:    NewConfigLogger(),
	}
}

// LoadAndValidate loads configuration from file and validates it
func (cm *ConfigManager) LoadAndValidate(serviceName string, config interface{}, validator ValidatorFunc) error {
	// Load configuration using viper
	v, err := cm.loader.LoadConfig(serviceName)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Unmarshal into the provided config struct
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate if validator is provided
	if validator != nil {
		if err := validator(config); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}

	return nil
}

// LoadConfigWithViper loads configuration and returns the viper instance
func (cm *ConfigManager) LoadConfigWithViper(serviceName string) (*viper.Viper, error) {
	return cm.loader.LoadConfig(serviceName)
}

// ValidateConfig runs validation on a configuration struct
func (cm *ConfigManager) ValidateConfig(config interface{}, validator ValidatorFunc) error {
	if validator != nil {
		return validator(config)
	}
	return nil
}

// LogConfig logs a configuration struct
func (cm *ConfigManager) LogConfig(config interface{}, sections map[string][]string) {
	cm.logger.LogConfig(config, sections)
}

// GetLogger returns the configuration logger
func (cm *ConfigManager) GetLogger() *ConfigLogger {
	return cm.logger
}

// Helper function for common service configuration pattern
func LoadServiceConfig(serviceName string, config interface{}, validator ValidatorFunc, logFunc func(*ConfigLogger, interface{})) error {
	manager := NewConfigManager()
	
	// Load and validate
	if err := manager.LoadAndValidate(serviceName, config, validator); err != nil {
		return err
	}
	
	// Log configuration
	if logFunc != nil {
		logFunc(manager.GetLogger(), config)
	}
	
	return nil
} 