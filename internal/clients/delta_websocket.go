package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Cryptovate-India/websocket-service/internal/config"
	"github.com/gorilla/websocket"
)

// MessageHandler is a function that handles messages from Delta Exchange
type MessageHandler func(message []byte, channelName string, productId string)

// DeltaWebsocketClient is a client for the Delta Exchange websocket API
type DeltaWebsocketClient struct {
	conn            *websocket.Conn
	url             string
	channels        []string
	productIDs      []string
	handlers        map[string]MessageHandler
	handlersMu      sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	reconnectMax    int
	reconnectCount  int
	reconnectDelay  time.Duration
	connected       bool
	connectedAt     time.Time
	lastError       string
	lastErrorAt     time.Time
	mu              sync.RWMutex
	subscriptions   map[string][]string // channel -> []assets
	subscriptionsMu sync.RWMutex
	totalMessages   int
}

// NewDeltaWebsocketClient creates a new Delta Exchange websocket client
func NewDeltaWebsocketClient(ctx context.Context, cfg *config.Delta) *DeltaWebsocketClient {
	clientCtx, cancel := context.WithCancel(ctx)

	client := &DeltaWebsocketClient{
		url:            cfg.URL,
		channels:       cfg.Channels,
		productIDs:     cfg.ProductIDs,
		handlers:       make(map[string]MessageHandler),
		ctx:            clientCtx,
		cancel:         cancel,
		reconnectMax:   cfg.ReconnectMax,
		reconnectDelay: 5 * time.Second,
		subscriptions:  make(map[string][]string),
	}

	return client
}

// Connect connects to the Delta Exchange websocket API
func (c *DeltaWebsocketClient) Connect() error {
	fmt.Println("Delta_WS: Connect: Attempting to connect to Delta Exchange...")

	// First check if already connected (with lock)
	c.mu.Lock()
	if c.connected {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	fmt.Println("Delta_WS: Connect: connecting url:", c.url)
	// Connect to the websocket (without holding the lock)
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		c.mu.Lock()
		c.lastError = fmt.Sprintf("Failed to connect to Delta Exchange: %v", err)
		c.lastErrorAt = time.Now()
		c.mu.Unlock()
		return fmt.Errorf("failed to connect to Delta Exchange: %w", err)
	}

	// Update connection state (with lock)
	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.connectedAt = time.Now()
	c.reconnectCount = 0
	c.mu.Unlock()

	fmt.Printf("Delta_WS: Connect: Successfully connected to %s\n", c.url)

	// Start the read pump
	go c.readPump()

	// Only restore existing subscriptions on reconnection (no automatic initial subscriptions)
	c.subscriptionsMu.RLock()
	hasExistingSubscriptions := len(c.subscriptions) > 0
	currentSubscriptions := make(map[string][]string)
	for channel, assets := range c.subscriptions {
		currentSubscriptions[channel] = append([]string{}, assets...)
	}
	c.subscriptionsMu.RUnlock()
	
	if hasExistingSubscriptions {
		// Reconnection - restore existing subscriptions
		fmt.Println("Delta_WS: Connect: Restoring existing subscriptions...")
		for channel, assets := range currentSubscriptions {
			fmt.Printf("Delta_WS: Connect: Restoring subscription to channel %s with assets: %v\n", channel, assets)
			if err := c.resubscribeChannel(channel, assets); err != nil {
				log.Printf("Failed to restore subscription to channel %s: %v", channel, err)
			}
		}
	} else {
		fmt.Println("Delta_WS: Connect: No existing subscriptions to restore. Waiting for client subscriptions...")
	}

	return nil
}

// RegisterHandler registers a handler for a channel
func (c *DeltaWebsocketClient) RegisterHandler(channel string, handler MessageHandler) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	c.handlers[channel] = handler
}

// Helper functions for asset management
func (c *DeltaWebsocketClient) containsAsset(assets []string, asset string) bool {
	for _, a := range assets {
		if a == asset {
			return true
		}
	}
	return false
}

func (c *DeltaWebsocketClient) addAsset(assets []string, asset string) []string {
	if !c.containsAsset(assets, asset) {
		assets = append(assets, asset)
	}
	return assets
}

func (c *DeltaWebsocketClient) removeAsset(assets []string, asset string) []string {
	for i, a := range assets {
		if a == asset {
			return append(assets[:i], assets[i+1:]...)
		}
	}
	return assets
}

// readPump reads messages from the websocket
func (c *DeltaWebsocketClient) readPump() {
	// Store a local reference to the connection to avoid race conditions
	var conn *websocket.Conn
	c.mu.RLock()
	conn = c.conn
	c.mu.RUnlock()

	if conn == nil {
		log.Println("Delta_WS: readPump: Connection is nil, cannot start read pump")
		return
	}

	defer func() {
		// Mark as disconnected
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()

		// Close the connection without holding the lock
		if conn != nil {
			conn.Close()
		}

		// Attempt to reconnect if the context is not canceled
		select {
		case <-c.ctx.Done():
			return
		default:
			c.reconnect()
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		
		fmt.Println("Delta_WS: readPump: Read message:", string(message))
		if err != nil {
			// Update error state with lock
			c.mu.Lock()
			c.lastError = fmt.Sprintf("Error reading from Delta Exchange: %v", err)
			c.lastErrorAt = time.Now()
			c.mu.Unlock()
			return
		}

		// Parse the message to get the channel
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error parsing message from Delta Exchange: %v", err)
			continue
		}

		// Check for the new message format with payload.channels
		var channel string
		var msgProductID string
		
		// Handle different message types
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case "subscriptions":
				// Handle subscription confirmation messages - these are just confirmations, not data
				fmt.Printf("Delta_WS: readPump: Subscription confirmation: %v\n", msg)
				continue
			default:
			channel = msgType
			}
		} else {
			log.Printf("Delta_WS: readPump: msg does not contain a valid 'type' field: %v", msg)
			continue
		}
		
		// Get the product ID from the symbol field
		if productId, ok := msg["symbol"].(string); ok {
			msgProductID = productId
		}

		// Check if this asset is actually subscribed before calling handler
		c.subscriptionsMu.RLock()
		subscribedAssets, channelExists := c.subscriptions[channel]
		isAssetSubscribed := false
		if channelExists && msgProductID != "" {
			isAssetSubscribed = c.containsAsset(subscribedAssets, msgProductID)
		} else if channelExists && msgProductID == "" {
			// If no specific product ID, assume it's a general channel message
			isAssetSubscribed = true
		}
		c.subscriptionsMu.RUnlock()

		// Only process message if the asset is subscribed or if it's a general channel message
		if !channelExists {
			fmt.Printf("Delta_WS: readPump: Ignoring message - channel '%s' not subscribed\n", channel)
			continue
		}

		if msgProductID != "" && !isAssetSubscribed {
			fmt.Printf("Delta_WS: readPump: Ignoring message - asset '%s' not subscribed to channel '%s'. Subscribed assets: %v\n", 
				msgProductID, channel, subscribedAssets)
			continue
		}

		// Call the handler for the channel
		c.handlersMu.RLock()
		c.totalMessages++
		fmt.Println("Delta_WS: readPump: Processing message - Channel:", channel, "product:", msgProductID, "Message count:", c.totalMessages)
		fmt.Printf("Delta_WS: readPump: Subscribed assets for channel '%s': %v\n", channel, subscribedAssets)
		
		// Find and call the handler for this channel
		handler, ok := c.handlers[channel]
		c.handlersMu.RUnlock()
		if ok {
			fmt.Printf("Delta_WS: readPump: Calling handler for channel '%s', asset '%s'\n", channel, msgProductID)
			handler(message, channel, msgProductID)
		} else {
			fmt.Printf("Delta_WS: readPump: No handler registered for channel '%s'\n", channel)
		}
	}
}

// subscribe subscribes to a channel
func (c *DeltaWebsocketClient) Subscribe(channel string, productIDs []string) error {
	// Use the product IDs as symbols
	symbols := []string{"all"}
	if len(productIDs) > 0 {
		symbols = productIDs
	}

	// Update internal subscription state first
	c.subscriptionsMu.Lock()
	var allAssets []string
	if existingAssets, exists := c.subscriptions[channel]; exists {
		// Add new assets to existing subscription
		allAssets = append([]string{}, existingAssets...)
		log.Println("Delta_WS: Subscribe: Existing assets:", allAssets)
		for _, asset := range symbols {
			allAssets = c.addAsset(allAssets, asset)
		}
		c.subscriptions[channel] = allAssets
	} else {
		// New channel subscription
		allAssets = append([]string{}, symbols...)
		log.Println("Delta_WS: Subscribe: New assets:", allAssets)
		c.subscriptions[channel] = allAssets
	}
	c.subscriptionsMu.Unlock()

	// Create the subscription message with ALL assets for the channel
	msg := map[string]interface{}{
		"type": "subscribe",
		"payload": map[string]interface{}{
			"channels": []map[string]interface{}{
				{
					"name":    channel,
					"symbols": allAssets,
				},
			},
		},
	}

	// Marshal the message
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Delta_WS: Subscribe: Error marshalling subscription message:", err)
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	// Check connection status and send message (with lock held)
	c.mu.RLock()
	if !c.connected || c.conn == nil {
		c.mu.RUnlock()
		return fmt.Errorf("not connected to Delta Exchange")
	}
	// Send the message while holding the lock to prevent TOCTOU race
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		c.mu.RUnlock()
		return fmt.Errorf("failed to send subscription message: %w", err)
	}
	c.mu.RUnlock()

	return nil
}

// unsubscribe unsubscribes from a channel
func (c *DeltaWebsocketClient) Unsubscribe(channel string) error {
	// Create the unsubscription message
	msg := map[string]interface{}{
		"type": "unsubscribe",
		"payload": map[string]interface{}{
			"channels": []map[string]interface{}{
				{
					"name": channel,
				},
			},
		},
	}
	// Marshal the message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal unsubscription message: %w", err)
	}
	// Check connection status and send message (with lock held)
	c.mu.RLock()
	if !c.connected || c.conn == nil {
		c.mu.RUnlock()
		return fmt.Errorf("not connected to Delta Exchange")
	}
	// Send the message while holding the lock to prevent TOCTOU race
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		c.mu.RUnlock()
		return fmt.Errorf("failed to send unsubscription message: %w", err)
	}
	c.mu.RUnlock()
	// Remove the subscription
	c.subscriptionsMu.Lock()
	delete(c.subscriptions, channel)
	c.subscriptionsMu.Unlock()
	return nil
}

// SubscribeAssets subscribes to specific assets on a channel
func (c *DeltaWebsocketClient) SubscribeAssets(channel string, assets []string) error {
	return c.Subscribe(channel, assets)
}

// UnsubscribeAssets unsubscribes from specific assets on a channel
func (c *DeltaWebsocketClient) UnsubscribeAssets(channel string, assetsToRemove []string) error {
	c.subscriptionsMu.Lock()
	currentAssets, exists := c.subscriptions[channel]
	if !exists {
		c.subscriptionsMu.Unlock()
		return fmt.Errorf("channel %s is not subscribed", channel)
	}
	
	// Remove specified assets
	for _, asset := range assetsToRemove {
		currentAssets = c.removeAsset(currentAssets, asset)
	}
	
	// Update subscription state
	if len(currentAssets) == 0 {
		// No assets left, remove entire channel
		delete(c.subscriptions, channel)
		c.subscriptionsMu.Unlock()
		return c.Unsubscribe(channel)
	} else {
		// Update with remaining assets
		c.subscriptions[channel] = currentAssets
		c.subscriptionsMu.Unlock()
		
		// Re-subscribe with remaining assets
		return c.resubscribeChannel(channel, currentAssets)
	}
}

// resubscribeChannel re-subscribes to a channel with specific assets
func (c *DeltaWebsocketClient) resubscribeChannel(channel string, assets []string) error {
	// Create the subscription message
	msg := map[string]interface{}{
		"type": "subscribe",
		"payload": map[string]interface{}{
			"channels": []map[string]interface{}{
				{
					"name":    channel,
					"symbols": assets,
				},
			},
		},
	}

	// Marshal the message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal resubscription message: %w", err)
	}

	// Check connection status and send message
	c.mu.RLock()
	if !c.connected || c.conn == nil {
		c.mu.RUnlock()
		return fmt.Errorf("not connected to Delta Exchange")
	}
	
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		c.mu.RUnlock()
		return fmt.Errorf("failed to send resubscription message: %w", err)
	}
	c.mu.RUnlock()

	return nil
}

// reconnect attempts to reconnect to the Delta Exchange websocket API
func (c *DeltaWebsocketClient) reconnect() {
	// Get the reconnect count with lock
	c.mu.Lock()
	c.reconnectCount++
	reconnectCount := c.reconnectCount
	reconnectMax := c.reconnectMax
	c.mu.Unlock()

	if reconnectCount > reconnectMax {
		log.Printf("Exceeded maximum reconnection attempts (%d)", reconnectMax)
		return
	}

	log.Printf("Reconnecting to Delta Exchange (attempt %d/%d)...", reconnectCount, reconnectMax)

	// Sleep without holding any locks
	time.Sleep(c.reconnectDelay)

	// Connect without holding any locks
	if err := c.Connect(); err != nil {
		log.Printf("Failed to reconnect to Delta Exchange: %v", err)
	}
}

// Close closes the connection to the Delta Exchange websocket API
func (c *DeltaWebsocketClient) Close() error {
	// Cancel the context first to signal all goroutines to stop
	c.cancel()

	// Get the connection with lock
	var conn *websocket.Conn
	c.mu.Lock()
	conn = c.conn
	c.connected = false
	c.conn = nil // Clear the connection reference
	c.mu.Unlock()

	// Close the connection without holding the lock
	if conn != nil {
		return conn.Close()
	}

	return nil
}

// IsConnected returns whether the client is connected to the Delta Exchange websocket API
func (c *DeltaWebsocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetConnectionStatus returns the connection status of the Delta Exchange websocket client
func (c *DeltaWebsocketClient) GetConnectionStatus() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := map[string]interface{}{
		"connected":           c.connected,
		"connected_at":        c.connectedAt,
		"reconnect_count":     c.reconnectCount,
		"reconnect_max":       c.reconnectMax,
		"last_error":          c.lastError,
		"last_error_at":       c.lastErrorAt,
		"subscribed_channels": c.getSubscribedChannels(),
	}

	return status
}

// getSubscribedChannels returns the channels that the client is subscribed to
func (c *DeltaWebsocketClient) getSubscribedChannels() []string {
	c.subscriptionsMu.RLock()
	defer c.subscriptionsMu.RUnlock()

	channels := make([]string, 0, len(c.subscriptions))
	for channel := range c.subscriptions {
		channels = append(channels, channel)
	}

	fmt.Println("Delta_WS: getSubscribedChannels: Subscribed channels:", channels)
	return channels
}

// IsAssetSubscribed checks if a specific asset is subscribed to a channel
func (c *DeltaWebsocketClient) IsAssetSubscribed(channel string, asset string) bool {
	c.subscriptionsMu.RLock()
	defer c.subscriptionsMu.RUnlock()
	
	if assets, exists := c.subscriptions[channel]; exists {
		return c.containsAsset(assets, asset)
	}
	return false
}

// GetSubscribedAssets returns all assets subscribed to a specific channel
func (c *DeltaWebsocketClient) GetSubscribedAssets(channel string) []string {
	c.subscriptionsMu.RLock()
	defer c.subscriptionsMu.RUnlock()
	
	if assets, exists := c.subscriptions[channel]; exists {
		// Return a copy to avoid external modification
		result := make([]string, len(assets))
		copy(result, assets)
		return result
	}
	return []string{}
}

// GetAllSubscriptions returns all current subscriptions
func (c *DeltaWebsocketClient) GetAllSubscriptions() map[string][]string {
	c.subscriptionsMu.RLock()
	defer c.subscriptionsMu.RUnlock()
	
	// Return a deep copy to avoid external modification
	result := make(map[string][]string)
	for channel, assets := range c.subscriptions {
		assetsCopy := make([]string, len(assets))
		copy(assetsCopy, assets)
		result[channel] = assetsCopy
	}
	return result
}
