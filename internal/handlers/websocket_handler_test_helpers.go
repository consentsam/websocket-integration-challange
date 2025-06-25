package handlers

import "github.com/gorilla/websocket"

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

// NewTestClient creates a dummy client for use in unit tests. The underlying
// websocket connection is nil but satisfies the handler logic.
func NewTestClient() *Client {
	return &Client{
		conn:           &websocket.Conn{},
		send:           make(chan []byte),
		subscriptions:  make(map[string]bool),
		productFilters: make(map[string][]string),
	}
}
