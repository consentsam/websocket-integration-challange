package bugs

import (
	"os"
	"path/filepath"
	"testing"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/consentsam/websocket-integration-challange/internal/config"
)

// loadConfigWithAbsolutePath loads config using absolute paths to avoid race conditions
func loadConfigWithAbsolutePath(serviceName string, projectRoot string) (*config.Config, error) {
	// Detect environment
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local"
	}

	// Set default configuration values (copied from config.LoadConfig)
	cfg := &config.Config{
		ServiceName: serviceName,
		Environment: environment,
		LogLevel:    "info",
		HTTPPort:    8083,
		GRPCPort:    9093,
	}

	// Set default metrics configuration
	cfg.Metrics.Enabled = true
	cfg.Metrics.Endpoint = "/metrics"

	// Initialize Viper with absolute paths
	v := viper.New()
	v.SetConfigName(environment)
	v.SetConfigType("yaml")
	
	// Use absolute path to config directory
	configDir := filepath.Join(projectRoot, "config")
	v.AddConfigPath(configDir)
	v.AddConfigPath(projectRoot)

	// Set environment variable support
	v.SetEnvPrefix("WEBSOCKET")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read the configuration file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, continue with defaults
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal the configuration
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func TestBug05_Repro(t *testing.T) {
	os.Setenv("ENVIRONMENT", "local")
	defer os.Unsetenv("ENVIRONMENT")

	// Get the current working directory and calculate project root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	// Calculate the project root path (go up two levels from tests/bugs/)
	projectRoot, err := filepath.Abs(filepath.Join(cwd, "..", ".."))
	if err != nil {
		t.Fatalf("failed to resolve absolute path to project root: %v", err)
	}

	// Load config using absolute paths - no directory changes needed!
	cfg, err := loadConfigWithAbsolutePath("websocket-service", projectRoot)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.HTTPPort != 8080 {
		t.Fatalf("expected HTTP port 8080, got %d", cfg.HTTPPort)
	}
	if cfg.GRPCPort != 9090 {
		t.Fatalf("expected gRPC port 9090, got %d", cfg.GRPCPort)
	}
}
