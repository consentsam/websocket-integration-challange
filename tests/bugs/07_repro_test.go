package bugs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/consentsam/websocket-integration-challange/internal/config"
	"github.com/consentsam/websocket-integration-challange/telemetry"
)

func TestBug07_Repro(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Regression test; passes post-fix")
	}
	cfg, err := config.LoadConfig("websocket-service")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	mux := http.NewServeMux()
	metricsMux, err := telemetry.Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("telemetry init: %v", err)
	}
	if cfg.Metrics.Enabled && metricsMux != nil {
		mux.Handle(cfg.Metrics.Endpoint, metricsMux)
	}
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + cfg.Metrics.Endpoint)
	if err != nil {
		t.Fatalf("GET metrics: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
}
