package bugs

import (
	"context"
	"testing"

	"github.com/consentsam/websocket-integration-challange/internal/clients"
	"github.com/consentsam/websocket-integration-challange/internal/config"
	handlers "github.com/consentsam/websocket-integration-challange/internal/handlers"
)

type mockDeltaClient struct{ unsubscribed []string }

func (m *mockDeltaClient) Connect() error                                      { return nil }
func (m *mockDeltaClient) Close() error                                        { return nil }
func (m *mockDeltaClient) Subscribe(channel string, productIDs []string) error { return nil }
func (m *mockDeltaClient) Unsubscribe(channel string) error {
	m.unsubscribed = append(m.unsubscribed, channel)
	return nil
}
func (m *mockDeltaClient) RegisterHandler(channel string, handler clients.MessageHandler) {}
func (m *mockDeltaClient) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{}
}
func (m *mockDeltaClient) IsConnected() bool { return true }

// TestBug12_Repro ensures the handler unsubscribes from Delta when the last client leaves a channel.
func TestBug12_Repro(t *testing.T) {
	cfg, err := config.LoadConfig("websocket-service")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	cfg.Delta.Enabled = false

	h := handlers.NewWebsocketHandler(context.Background(), cfg)
	mock := &mockDeltaClient{}
	h.SetDeltaClient(mock)

	c1 := handlers.NewTestClient()
	c2 := handlers.NewTestClient()

	h.TestSubscribeClient(c1, "test-channel", []string{"all"})
	h.TestSubscribeClient(c2, "test-channel", []string{"all"})

	if cnt := h.TestSubscriptionsCount("test-channel"); cnt != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", cnt)
	}

	msg := map[string]interface{}{
		"type": "unsubscribe",
		"payload": map[string]interface{}{
			"channels": []interface{}{map[string]interface{}{"name": "test-channel"}},
		},
	}

	h.TestHandleUnsubscribe(c1, msg)
	if len(mock.unsubscribed) != 0 {
		t.Fatalf("delta unsubscribed too early: %v", mock.unsubscribed)
	}

	h.TestHandleUnsubscribe(c2, msg)
	if len(mock.unsubscribed) != 1 || mock.unsubscribed[0] != "test-channel" {
		t.Fatalf("delta unsubscribe not called correctly: %v", mock.unsubscribed)
	}

	if cnt := h.TestSubscriptionsCount("test-channel"); cnt != 0 {
		t.Fatalf("subscriptions not cleaned up, got %d", cnt)
	}
}
