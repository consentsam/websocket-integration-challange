package bugs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/consentsam/websocket-integration-challange/internal/clients"
	"github.com/consentsam/websocket-integration-challange/internal/config"
)

// TestBug08_Repro verifies that Connect automatically subscribes to channels.
func TestBug08_Repro(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}

	// Channel to capture the first message from client
	msgCh := make(chan []byte, 1)

	// Start a local websocket server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := up.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade error: %v", err)
			return
		}
		defer conn.Close()

		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("read message error: %v", err)
			return
		}
		msgCh <- msg
		// Keep connection open briefly to avoid client errors
		time.Sleep(100 * time.Millisecond)
	}))
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"):] // convert http://127.0.0.1 -> ws://127.0.0.1

	cfg := &config.Delta{
		Enabled:      true,
		URL:          wsURL,
		Channels:     []string{"v2/ticker"},
		ProductIDs:   []string{"BTCUSD"},
		ReconnectMax: 1,
	}

	client := clients.NewDeltaWebsocketClient(context.Background(), cfg)

	if err := client.Connect(); err != nil {
		t.Fatalf("connect error: %v", err)
	}

	select {
	case raw := <-msgCh:
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if m["type"] != "subscribe" {
			t.Fatalf("expected subscribe message, got %v", m["type"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no message received")
	}

	client.Close()
}
