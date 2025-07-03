package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ConfigLoader defines the interface for loading configuration
type ConfigLoader interface {
	LoadConfig(serviceName string) (*viper.Viper, error)
	SetDefaults(v *viper.Viper, serviceName string)
	GetEnvironment() string
}

// ViperConfigLoader implements ConfigLoader using Viper
type ViperConfigLoader struct {
	configPaths []string
	defaults    map[string]interface{}
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() ConfigLoader {
	return &ViperConfigLoader{
		configPaths: []string{
			"./config",
			"../config",
			"../../config",
			".",
		},
		defaults: make(map[string]interface{}),
	}
}

// LoadConfig loads the configuration from YAML files
func (l *ViperConfigLoader) LoadConfig(serviceName string) (*viper.Viper, error) {
	// Initialize viper instance
	v := viper.New()

	// Determine the environment to use
	env := l.GetEnvironment()

	// Set up viper configuration
	v.SetConfigName(env)
	v.SetConfigType("yaml")
	
	// Add all config paths
	for _, path := range l.configPaths {
		v.AddConfigPath(path)
	}

	// Enable automatic environment variable reading
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default configuration values
	l.SetDefaults(v, serviceName)

	// Read the configuration file
	configFile := filepath.Join("config", fmt.Sprintf("%s.yaml", env))
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults with a warning
			log.Printf("Warning: Config file '%s' not found, using default values", configFile)
		} else {
			return nil, fmt.Errorf("error reading config file '%s': %w", configFile, err)
		}
	} else {
		log.Printf("Successfully loaded config file: %s", v.ConfigFileUsed())
	}

	return v, nil
}

// GetEnvironment returns the current environment
func (l *ViperConfigLoader) GetEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local" // Default to local environment
	}
	return env
}

// SetDefaults sets default configuration values
func (l *ViperConfigLoader) SetDefaults(v *viper.Viper, serviceName string) {
	// Service defaults
	v.SetDefault("service_name", serviceName)
	v.SetDefault("environment", l.GetEnvironment())
	v.SetDefault("log_level", "info")
	v.SetDefault("http_port", 8080)
	v.SetDefault("grpc_port", 9090)

	// Websocket defaults
	v.SetDefault("websocket.read_buffer_size", 1024)
	v.SetDefault("websocket.write_buffer_size", 1024)
	v.SetDefault("websocket.max_message_size", 4096)
	v.SetDefault("websocket.check_origin", false)
	v.SetDefault("websocket.auth.required", false)
	v.SetDefault("websocket.auth.secret", "default-secret")

	// Security defaults
	v.SetDefault("security.cors_enabled", true)
	v.SetDefault("security.cors_allowed_origins", "*")
	v.SetDefault("security.rate_limit_enabled", false)
	v.SetDefault("security.rate_limit_requests", 100)
	v.SetDefault("security.rate_limit_duration", 60)

	// Delta Exchange defaults
	v.SetDefault("delta.enabled", true)
	v.SetDefault("delta.url", "wss://socket.india.delta.exchange")
	v.SetDefault("delta.channels", []string{"v2/ticker"})
	v.SetDefault("delta.product_ids", []string{"BTC_USDT"})
	v.SetDefault("delta.reconnect_max", 5)

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.endpoint", "/metrics")
} 