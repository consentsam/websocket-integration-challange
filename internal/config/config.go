package config

import (
	"fmt"

	sc "github.com/Cryptovate-India/server-utils/config"
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

// LoadConfig loads the configuration from the config file using server-utils
func LoadConfig(serviceName string) (*Config, error) {
	config := &Config{}
	
	// Use the simplified LoadServiceConfig helper
	err := sc.LoadServiceConfig(
		serviceName,
		config,
		buildValidator(),
		logConfig,
	)
	
	if err != nil {
		return nil, err
	}
	
	return config, nil
}

// buildValidator creates the validation function for the config
func buildValidator() sc.ValidatorFunc {
	return sc.CombineValidators(
		// Service validation
		validateService,
		// Websocket validation
		validateWebsocket,
		// Delta validation
		validateDelta,
		// Security validation
		validateSecurity,
	)
}

// validateService validates service configuration
func validateService(config interface{}) error {
	cfg := config.(*Config)
	
	if err := sc.ValidateNonEmpty(cfg.ServiceName, "service_name"); err != nil {
		return err
	}
	if err := sc.ValidatePortRange(cfg.HTTPPort, "http_port"); err != nil {
		return err
	}
	if err := sc.ValidatePortRange(cfg.GRPCPort, "grpc_port"); err != nil {
		return err
	}
	if cfg.HTTPPort == cfg.GRPCPort {
		return fmt.Errorf("http_port and grpc_port cannot be the same (%d)", cfg.HTTPPort)
	}
	return nil
}

// validateWebsocket validates websocket configuration
func validateWebsocket(config interface{}) error {
	cfg := config.(*Config)
	
	if err := sc.ValidatePositive(cfg.Websocket.ReadBufferSize, "websocket.read_buffer_size"); err != nil {
		return err
	}
	if err := sc.ValidatePositive(cfg.Websocket.WriteBufferSize, "websocket.write_buffer_size"); err != nil {
		return err
	}
	if err := sc.ValidatePositiveInt64(cfg.Websocket.MaxMessageSize, "websocket.max_message_size"); err != nil {
		return err
	}
	return nil
}

// validateDelta validates Delta Exchange configuration
func validateDelta(config interface{}) error {
	cfg := config.(*Config)
	
	if cfg.Delta.Enabled {
		if err := sc.ValidateNonEmpty(cfg.Delta.URL, "delta.url"); err != nil {
			return err
		}
		if err := sc.ValidateNonEmptySlice(cfg.Delta.Channels, "delta.channels"); err != nil {
			return err
		}
		// Removed validation for product_ids - allow empty to use default behavior
		if err := sc.ValidateNonNegative(cfg.Delta.ReconnectMax, "delta.reconnect_max"); err != nil {
			return err
		}
	}
	return nil
}

// validateSecurity validates security configuration
func validateSecurity(config interface{}) error {
	cfg := config.(*Config)
	
	if cfg.Security.RateLimitEnabled {
		if err := sc.ValidatePositive(cfg.Security.RateLimitRequests, "security.rate_limit_requests"); err != nil {
			return err
		}
		if err := sc.ValidatePositive(cfg.Security.RateLimitDuration, "security.rate_limit_duration"); err != nil {
			return err
		}
	}
	return nil
}

// logConfig logs the configuration using server-utils helpers
func logConfig(logger *sc.ConfigLogger, config interface{}) {
	cfg := config.(*Config)
	
	// Log service configuration
	logger.LogServiceConfig(
		cfg.ServiceName,
		cfg.Environment,
		cfg.LogLevel,
		cfg.HTTPPort,
		cfg.GRPCPort,
	)
	
	// Log websocket configuration
	logger.LogWebsocketConfig(
		cfg.Websocket.ReadBufferSize,
		cfg.Websocket.WriteBufferSize,
		cfg.Websocket.MaxMessageSize,
		cfg.Websocket.CheckOrigin,
		false, // auth_required - always false (no auth)
		"",    // auth_secret - always empty (no auth)
	)
	
	// Log security configuration
	logger.LogSecurityConfig(
		cfg.Security.CORSEnabled,
		cfg.Security.CORSAllowedOrigins,
		cfg.Security.RateLimitEnabled,
		cfg.Security.RateLimitRequests,
		cfg.Security.RateLimitDuration,
	)
	
	// Log Delta configuration
	logger.LogDeltaConfig(
		cfg.Delta.Enabled,
		cfg.Delta.URL,
		cfg.Delta.Channels,
		cfg.Delta.ProductIDs,
		cfg.Delta.ReconnectMax,
	)
	
	// Log metrics configuration
	logger.LogMetricsConfig(
		cfg.Metrics.Enabled,
		cfg.Metrics.Endpoint,
	)
}

// GetCORSAllowedOrigins returns the allowed origins for CORS
func (c *Config) GetCORSAllowedOrigins() []string {
	return sc.ParseCSV(c.Security.CORSAllowedOrigins)
}
