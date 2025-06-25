package bugs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/consentsam/websocket-integration-challange/internal/config"
)

// TestBug05_Repro ensures that config.LoadConfig correctly reads the YAML configuration
// file for the detected environment (defaulting to "local") and that the HTTP/GRPC
// port values can be overridden via the HTTP_PORT and GRPC_PORT environment variables.
func TestBug05_Repro(t *testing.T) {
	// Force the ENVIRONMENT to "local" so that local.yaml is chosen.
	t.Setenv("ENVIRONMENT", "local")

	// Determine project root so we can chdir there – config.LoadConfig expects the
	// ./config directory to be relative to the working directory.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	projectRoot, err := filepath.Abs(filepath.Join(cwd, "..", ".."))
	if err != nil {
		t.Fatalf("failed to resolve absolute path to project root: %v", err)
	}

	// Change to the project root for the duration of this test so the config loader
	// can locate ./config/local.yaml regardless of where the test was started from.
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to capture working directory: %v", err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("failed to chdir to project root: %v", err)
	}
	defer os.Chdir(originalWd)

	// 1. Verify values loaded from YAML without overrides.
	cfg, err := config.LoadConfig("websocket-service")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.HTTPPort != 8080 {
		t.Fatalf("expected HTTP port 8080 from YAML, got %d", cfg.HTTPPort)
	}
	if cfg.GRPCPort != 9090 {
		t.Fatalf("expected gRPC port 9090 from YAML, got %d", cfg.GRPCPort)
	}
	if !cfg.Delta.Enabled {
		t.Fatalf("expected Delta section to be initialised; got %+v", cfg.Delta)
	}

	// 2. Verify environment variable overrides take precedence.
	t.Setenv("HTTP_PORT", "8888")
	t.Setenv("GRPC_PORT", "9999")

	cfg, err = config.LoadConfig("websocket-service")
	if err != nil {
		t.Fatalf("failed to load config with overrides: %v", err)
	}
	if cfg.HTTPPort != 8888 {
		t.Fatalf("expected HTTP port override 8888, got %d", cfg.HTTPPort)
	}
	if cfg.GRPCPort != 9999 {
		t.Fatalf("expected gRPC port override 9999, got %d", cfg.GRPCPort)
	}
}
