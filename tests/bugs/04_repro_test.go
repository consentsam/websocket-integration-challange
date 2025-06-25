package bugs

import (
	"context"
	"testing"

	"github.com/consentsam/websocket-integration-challange/internal/clients"
	"github.com/consentsam/websocket-integration-challange/internal/config"
	handlers "github.com/consentsam/websocket-integration-challange/internal/handlers"
)

type stubDeltaClient struct {
	unsubscribed []string
}

func (s *stubDeltaClient) Connect() error                                      { return nil }
func (s *stubDeltaClient) Close() error                                        { return nil }
func (s *stubDeltaClient) Subscribe(channel string, productIDs []string) error { return nil }
func (s *stubDeltaClient) Unsubscribe(channel string) error {
	s.unsubscribed = append(s.unsubscribed, channel)
	return nil
}
func (s *stubDeltaClient) RegisterHandler(channel string, handler clients.MessageHandler) {}
func (s *stubDeltaClient) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{}
}
func (s *stubDeltaClient) IsConnected() bool { return true }

func TestBug04_Repro(t *testing.T) {
	cfg := &config.Config{}
	cfg.Delta.Enabled = false
	h := handlers.NewWebsocketHandler(context.Background(), cfg)
	h.SetDeltaClient(&stubDeltaClient{})

	client := handlers.NewTestClient()

	h.TestSubscribeClient(client, "v2/ticker", []string{"all"})

	msg := map[string]interface{}{
		"type": "unsubscribe",
		"payload": map[string]interface{}{
			"channels": []interface{}{
				map[string]interface{}{"name": "v2/ticker"},
			},
		},
	}

	h.TestHandleUnsubscribe(client, msg)

	if count := h.TestSubscriptionsCount("v2/ticker"); count != 0 {
		t.Fatalf("subscription should be removed, got %d", count)
	}

	stub := h.GetDeltaClient().(*stubDeltaClient)
	if len(stub.unsubscribed) != 1 || stub.unsubscribed[0] != "v2/ticker" {
		t.Fatalf("delta unsubscribe not called correctly: %#v", stub.unsubscribed)
	}
}
