package handlers

import (
	"fmt"
	"sync"
	"time"

	"github.com/consentsam/websocket-integration-challange/internal/clients"
)

// Helper functions used in tests to interact with the handler without going
// through the network layer.

func (h *WebsocketHandler) RegisterClientForTest(c *Client) {
	h.register <- c
}

func (h *WebsocketHandler) UnregisterClientForTest(c *Client) {
	h.unregister <- c
}

func (h *WebsocketHandler) BroadcastForTest(msg []byte) {
	h.broadcast <- msg
}

// TestSubscribeClient exposes subscribeClient for tests.
func (h *WebsocketHandler) TestSubscribeClient(c *Client, channel string, productIDs []string) {
	h.subscribeClient(c, channel, productIDs)
}

// TestUnsubscribeClient exposes unsubscribeClient for tests.
func (h *WebsocketHandler) TestUnsubscribeClient(c *Client, channel string) {
	h.unsubscribeClient(c, channel)
}

// TestHandleUnsubscribe exposes handleUnsubscribe for tests.
func (h *WebsocketHandler) TestHandleUnsubscribe(c *Client, msg map[string]interface{}) {
	h.handleUnsubscribe(c, msg)
}

// TestSubscriptionsCount returns the number of clients subscribed to a channel.
func (h *WebsocketHandler) TestSubscriptionsCount(channel string) int {
	h.subscriptionsMu.RLock()
	defer h.subscriptionsMu.RUnlock()
	if clients, ok := h.subscriptions[channel]; ok {
		return len(clients)
	}
	return 0
}

// NewTestClient creates a comprehensive test Client that matches the production structure
// but doesn't require a real websocket connection.
func NewTestClient() *Client {
	now := time.Now()
	return &Client{
		conn:           nil,                    // No real websocket connection needed for tests
		send:           make(chan []byte, 256), // Match production buffer size
		subscriptions:  make(map[string]bool),
		productFilters: make(map[string][]string),
		mu:             sync.RWMutex{},
		id:             fmt.Sprintf("test-%d", now.UnixNano()),
		connectedAt:    now,
		lastActivity:   now,
	}
}

// GetDeltaClient exposes the current DeltaClient instance.
func (h *WebsocketHandler) GetDeltaClient() clients.DeltaClient {
	return h.deltaClient
}
