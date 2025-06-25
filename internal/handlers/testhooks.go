package handlers

import "github.com/consentsam/websocket-integration-challange/internal/clients"

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

// NewTestClient creates a minimal Client for unit tests.
func NewTestClient() *Client {
	return &Client{
		send:           make(chan []byte, 1),
		subscriptions:  make(map[string]bool),
		productFilters: make(map[string][]string),
	}
}

// GetDeltaClient exposes the current DeltaClient instance.
func (h *WebsocketHandler) GetDeltaClient() clients.DeltaClient {
	return h.deltaClient
}
