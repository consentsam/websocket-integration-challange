package bugs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/consentsam/websocket-integration-challange/internal/config"
	handlers "github.com/consentsam/websocket-integration-challange/internal/handlers"
)

// TestBug11_Repro verifies that messages are not batched into malformed JSON.
func TestBug11_Repro(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Regression test; passes post-fix")
	}

	cfg, err := config.LoadConfig("websocket-service")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	cfg.Delta.Enabled = false

	h := handlers.NewWebsocketHandler(context.Background(), cfg)
	defer h.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.HandleWebsocket)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"):] + "/ws"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer c.Close()

	sub := `{"type":"subscribe","payload":{"channels":[{"name":"ticker","symbols":["all"]}]}}`
	if err := c.WriteMessage(websocket.TextMessage, []byte(sub)); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if _, _, err := c.ReadMessage(); err != nil {
		t.Fatalf("read confirm: %v", err)
	}

	msg1 := []byte(`{"type":"message1","data":"test1"}`)
	msg2 := []byte(`{"type":"message2","data":"test2"}`)
	h.BroadcastToChannel("ticker", msg1, "all")
	h.BroadcastToChannel("ticker", msg2, "all")

	c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw1, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read first: %v", err)
	}
	var m1 map[string]any
	if err := json.Unmarshal(raw1, &m1); err != nil {
		t.Fatalf("invalid json1: %v", err)
	}

	_, raw2, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read second: %v", err)
	}
	var m2 map[string]any
	if err := json.Unmarshal(raw2, &m2); err != nil {
		t.Fatalf("invalid json2: %v", err)
	}
}
