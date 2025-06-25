package bugs

import (
	"context"
	"sync"
	"testing"

	"github.com/consentsam/websocket-integration-challange/internal/config"
	handlers "github.com/consentsam/websocket-integration-challange/internal/handlers"
)

// TestBug09_Repro triggers the race condition in BroadcastToChannel when
// subscriptions are modified concurrently.
func TestBug09_Repro(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadConfig("websocket-service")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	cfg.Delta.Enabled = false

	h := handlers.NewWebsocketHandler(ctx, cfg)

	clients := make([]*handlers.Client, 100)
	for i := range clients {
		c := handlers.NewTestClient()
		clients[i] = c
		h.RegisterClientForTest(c)
		h.TestSubscribeClient(c, "test", []string{"p1"})
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			h.BroadcastToChannel("test", []byte("m"), "p1")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			h.TestUnsubscribeClient(clients[i], "test")
		}
	}()

	wg.Wait()
}
