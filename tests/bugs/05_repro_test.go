package bugs

import (
	"os"
	"testing"

	"github.com/consentsam/websocket-integration-challange/internal/config"
)

func TestBug05_Repro(t *testing.T) {
	os.Setenv("ENVIRONMENT", "local")
	defer os.Unsetenv("ENVIRONMENT")

	// Ensure working directory is project root so config files are found
	cwd, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(cwd)

	cfg, err := config.LoadConfig("websocket-service")
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
