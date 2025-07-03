package config

import (
	"log"
	"reflect"
)

// ConfigLogger provides structured logging for configuration objects
type ConfigLogger struct {
	maskSecrets bool
}

// NewConfigLogger creates a new configuration logger
func NewConfigLogger() *ConfigLogger {
	return &ConfigLogger{
		maskSecrets: true,
	}
}

// LogConfig logs a configuration struct with proper formatting
func (cl *ConfigLogger) LogConfig(config interface{}, sections map[string][]string) {
	log.Printf("=== Configuration Loaded ===")
	
	// Log top-level fields first
	cl.logStructFields(config, 0, []string{})
	
	// Log sections in order if provided
	for sectionName, fieldPaths := range sections {
		LogConfigSection(sectionName)
		for _, fieldPath := range fieldPaths {
			cl.logFieldByPath(config, fieldPath, 1)
		}
	}
	
	log.Printf("=== End Configuration ===")
}

// LogServiceConfig provides a standard way to log service configuration
func (cl *ConfigLogger) LogServiceConfig(serviceName, environment, logLevel string, httpPort, grpcPort int) {
	LogConfigValue("Service", serviceName, 0)
	LogConfigValue("Environment", environment, 0)
	LogConfigValue("Log Level", logLevel, 0)
	LogConfigValue("HTTP Port", httpPort, 0)
	LogConfigValue("gRPC Port", grpcPort, 0)
}

// LogWebsocketConfig logs websocket-specific configuration
func (cl *ConfigLogger) LogWebsocketConfig(readBuffer, writeBuffer int, maxMessageSize int64, checkOrigin, authRequired bool, authSecret string) {
	LogConfigSection("Websocket")
	LogConfigValue("Read Buffer Size", readBuffer, 1)
	LogConfigValue("Write Buffer Size", writeBuffer, 1)
	LogConfigValue("Max Message Size", maxMessageSize, 1)
	LogConfigValue("Check Origin", checkOrigin, 1)
	LogConfigValue("Auth Required", authRequired, 1)
	if authSecret != "" && cl.maskSecrets {
		LogConfigValue("Auth Secret", MaskSecret(authSecret), 1)
	}
}

// LogSecurityConfig logs security-specific configuration
func (cl *ConfigLogger) LogSecurityConfig(corsEnabled bool, corsOrigins string, rateLimitEnabled bool, rateLimitRequests, rateLimitDuration int) {
	LogConfigSection("Security")
	LogConfigValue("CORS Enabled", corsEnabled, 1)
	LogConfigValue("CORS Allowed Origins", corsOrigins, 1)
	LogConfigValue("Rate Limit Enabled", rateLimitEnabled, 1)
	if rateLimitEnabled {
		LogConfigValue("Rate Limit Requests", rateLimitRequests, 1)
		LogConfigValue("Rate Limit Duration", rateLimitDuration, 1)
	}
}

// LogDeltaConfig logs Delta Exchange specific configuration
func (cl *ConfigLogger) LogDeltaConfig(enabled bool, url string, channels, productIDs []string, reconnectMax int) {
	LogConfigSection("Delta Exchange")
	LogConfigValue("Enabled", enabled, 1)
	if enabled {
		LogConfigValue("URL", url, 1)
		LogConfigValue("Channels", channels, 1)
		LogConfigValue("Product IDs", productIDs, 1)
		LogConfigValue("Reconnect Max", reconnectMax, 1)
	}
}

// LogMetricsConfig logs metrics-specific configuration
func (cl *ConfigLogger) LogMetricsConfig(enabled bool, endpoint string) {
	LogConfigSection("Metrics")
	LogConfigValue("Enabled", enabled, 1)
	if enabled {
		LogConfigValue("Endpoint", endpoint, 1)
	}
}

// Helper methods for reflection-based logging

func (cl *ConfigLogger) logStructFields(v interface{}, indent int, exclude []string) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return
	}
	
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Skip excluded fields
		if contains(exclude, field.Name) {
			continue
		}
		
		// Handle nested structs
		if fieldVal.Kind() == reflect.Struct {
			continue // Skip inline structs for top-level logging
		}
		
		// Log the field
		LogConfigValue(field.Name, fieldVal.Interface(), indent)
	}
}

func (cl *ConfigLogger) logFieldByPath(v interface{}, path string, indent int) {
	// This is a simplified version - can be extended to support nested paths
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	field := val.FieldByName(path)
	if field.IsValid() {
		LogConfigValue(path, field.Interface(), indent)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
} 