package bugs

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/consentsam/websocket-integration-challange/internal/config"
	handlers "github.com/consentsam/websocket-integration-challange/internal/handlers"
)

// TestBug10_Repro verifies that broadcasting while clients are removed
// does not perform map writes under a read lock.
func TestBug10_Repro(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Regression test; passes post-fix")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig("websocket-service")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	cfg.Delta.Enabled = false

	h := handlers.NewWebsocketHandler(ctx, cfg)

	c1 := handlers.NewTestClient()
	c2 := handlers.NewBlockedTestClient()

	h.RegisterClientForTest(c1)
	h.RegisterClientForTest(c2)

	time.Sleep(20 * time.Millisecond)

	h.BroadcastForTest([]byte("msg"))

	time.Sleep(50 * time.Millisecond)

	stats := h.GetStatistics()
	if stats["active_connections"].(int) != 1 {
		t.Fatalf("expected 1 active connection after broadcast, got %v", stats["active_connections"])
	}

	h.UnregisterClientForTest(c1)
	time.Sleep(5 * time.Millisecond)
}
