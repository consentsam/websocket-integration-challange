package bugs

import (
	"context"
	"testing"
	"time"

	"github.com/consentsam/websocket-integration-challange/internal/config"
	handlers "github.com/consentsam/websocket-integration-challange/internal/handlers"
)

// TestBug03_Repro triggers the race condition in the WebsocketHandler broadcast
// logic when the handler deletes a client from the map while holding only a
// read lock. The test intentionally runs with the race detector to surface the
// issue.
func TestBug03_Repro(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig("websocket-service")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	cfg.Delta.Enabled = false

	h := handlers.NewWebsocketHandler(ctx, cfg)

	c1 := handlers.NewTestClient()
	c2 := handlers.NewTestClient()

	h.RegisterClientForTest(c1)
	h.RegisterClientForTest(c2)

	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			h.GetStatistics()
			time.Sleep(1 * time.Millisecond)
		}
		close(done)
	}()

	h.BroadcastForTest([]byte("msg"))

	<-done

	h.UnregisterClientForTest(c1)
	h.UnregisterClientForTest(c2)
	time.Sleep(10 * time.Millisecond)
}
