package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Delta represents the configuration for the Delta Exchange websocket
type Delta struct {
	Enabled      bool     `mapstructure:"enabled"`
	URL          string   `mapstructure:"url"`
	Channels     []string `mapstructure:"channels"`
	ProductIDs   []string `mapstructure:"product_ids"`
	ReconnectMax int      `mapstructure:"reconnect_max"`
}

// Config represents the configuration for the websocket service
type Config struct {
	// Service configuration
	ServiceName string `mapstructure:"service_name"`
	Environment string `mapstructure:"environment"`
	LogLevel    string `mapstructure:"log_level"`
	HTTPPort    int    `mapstructure:"http_port"`
	GRPCPort    int    `mapstructure:"grpc_port"`

	// Websocket configuration
	Websocket struct {
		ReadBufferSize  int   `mapstructure:"read_buffer_size"`
		WriteBufferSize int   `mapstructure:"write_buffer_size"`
		MaxMessageSize  int64 `mapstructure:"max_message_size"`
		CheckOrigin     bool  `mapstructure:"check_origin"`
		Auth            struct {
			Required bool   `mapstructure:"required"`
			Secret   string `mapstructure:"secret"`
		} `mapstructure:"auth"`
	} `mapstructure:"websocket"`

	// Security configuration
	Security struct {
		CORSEnabled        bool   `mapstructure:"cors_enabled"`
		CORSAllowedOrigins string `mapstructure:"cors_allowed_origins"`
		RateLimitEnabled   bool   `mapstructure:"rate_limit_enabled"`
		RateLimitRequests  int    `mapstructure:"rate_limit_requests"`
		RateLimitDuration  int    `mapstructure:"rate_limit_duration"`
	} `mapstructure:"security"`

	// Delta Exchange configuration
	Delta Delta `mapstructure:"delta"`

	// Metrics configuration
	Metrics struct {
		Enabled  bool   `mapstructure:"enabled"`
		Endpoint string `mapstructure:"endpoint"`
	} `mapstructure:"metrics"`
}

// LoadConfig loads the configuration from the config file
func LoadConfig(serviceName string) (*Config, error) {
	// Detect environment
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local"
	}

	// Set default configuration values
	config := &Config{
		ServiceName: serviceName,
		Environment: environment,
		LogLevel:    "info",
		HTTPPort:    8083,
		GRPCPort:    9093,
		Delta: Delta{
			Enabled:      true,
			URL:          "wss://socket.india.delta.exchange",
			Channels:     []string{"v2/ticker"},
			ProductIDs:   []string{"BTCUSD"},
			ReconnectMax: 5,
		},
	}

	// Set default metrics configuration
	config.Metrics.Enabled = true
	config.Metrics.Endpoint = "/metrics"

	// Initialize Viper
	v := viper.New()
	v.SetConfigName(environment)
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Set environment variable support
	v.SetEnvPrefix("WEBSOCKET")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read the configuration file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Config file not found for environment '%s', using defaults", environment)
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		log.Printf("Using config file: %s", v.ConfigFileUsed())
	}

	// Unmarshal the configuration
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Update environment field to match what was detected/loaded
	config.Environment = environment

	// Log the configuration
	log.Printf("Loaded configuration for %s in %s environment", config.ServiceName, config.Environment)
	log.Printf("HTTP Port: %d, gRPC Port: %d", config.HTTPPort, config.GRPCPort)
	log.Printf("Metrics Enabled: %v, Endpoint: %s", config.Metrics.Enabled, config.Metrics.Endpoint)

	return config, nil
}

// GetCORSAllowedOrigins returns the allowed origins for CORS
func (c *Config) GetCORSAllowedOrigins() []string {
	if c.Security.CORSAllowedOrigins == "" {
		return []string{"*"}
	}
	return strings.Split(c.Security.CORSAllowedOrigins, ",")
}
