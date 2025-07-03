package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Cryptovate-India/websocket-service/internal/clients"
	"github.com/Cryptovate-India/websocket-service/internal/config"
	"github.com/gorilla/websocket"
)

// Client represents a connected websocket client
type Client struct {
	conn           *websocket.Conn
	send           chan []byte
	subscriptions  map[string]bool
	productFilters map[string][]string
	mu             sync.RWMutex
	id             string
	connectedAt    time.Time
	lastActivity   time.Time
}

// WebsocketHandler handles websocket connections
type WebsocketHandler struct {
	upgrader         websocket.Upgrader
	clients          map[*Client]bool
	clientsMu        sync.RWMutex
	broadcast        chan []byte
	register         chan *Client
	unregister       chan *Client
	subscriptions    map[string]map[*Client]bool
	subscriptionsMu  sync.RWMutex
	registeredHandlers map[string]bool // Track which channels have handlers registered
	handlersMu       sync.RWMutex
	config           *config.Config
	deltaClient      *clients.DeltaWebsocketClient
	ctx              context.Context
	cancel           context.CancelFunc
	messagesSent     int64
	messagesReceived int64
}

// NewWebsocketHandler creates a new websocket handler
func NewWebsocketHandler(ctx context.Context, cfg *config.Config) *WebsocketHandler {
	handlerCtx, cancel := context.WithCancel(ctx)

	// Create the websocket handler
	handler := &WebsocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.Websocket.ReadBufferSize,
			WriteBufferSize: cfg.Websocket.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				if !cfg.Websocket.CheckOrigin {
					return true
				}
				// Check the origin against the allowed origins
				origin := r.Header.Get("Origin")
				for _, allowedOrigin := range cfg.GetCORSAllowedOrigins() {
					if allowedOrigin == "*" || allowedOrigin == origin {
						return true
					}
				}
				return false
			},
		},
		clients:            make(map[*Client]bool),
		broadcast:          make(chan []byte),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		subscriptions:      make(map[string]map[*Client]bool),
		registeredHandlers: make(map[string]bool),
		config:             cfg,
		ctx:                handlerCtx,
		cancel:             cancel,
	}

	fmt.Println("Websocket handler created")
	fmt.Println("Websocket handler config:", cfg)
	fmt.Println("Websocket handler context:", handlerCtx)

	// Create the Delta Exchange client if enabled
	if cfg.Delta.Enabled {
		handler.deltaClient = clients.NewDeltaWebsocketClient(handlerCtx, &cfg.Delta)

		// Register handlers for Delta Exchange channels from config
		for _, channel := range cfg.Delta.Channels {
			handler.registerDeltaHandler(channel)
			handler.registeredHandlers[channel] = true
			fmt.Println("WS_handler:ctor: Registered initial handler for channel: ", channel)
		}

		fmt.Println("WS_handler:ctor: Delta Exchange client created")

		// Connect to Delta Exchange
		if err := handler.deltaClient.Connect(); err != nil {
			fmt.Println("Failed to connect to Delta Exchange: ", err)
		}
	}

	// Start the handler
	go handler.run()

	return handler
}

// HandleWebsocket handles a websocket connection
func (h *WebsocketHandler) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the connection to a websocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create a new client
	client := &Client{
		conn:           conn,
		send:           make(chan []byte, 256),
		subscriptions:  make(map[string]bool),
		productFilters: make(map[string][]string),
		id:             fmt.Sprintf("%d", time.Now().UnixNano()),
		connectedAt:    time.Now(),
		lastActivity:   time.Now(),
	}

	// Register the client
	h.register <- client

	// Start the read and write pumps
	go h.readPump(client)
	go h.writePump(client)
}

// BroadcastToChannel broadcasts a message to all clients subscribed to a channel
func (h *WebsocketHandler) BroadcastToChannel(channel string, message []byte, productID string) {
	// Get the clients subscribed to the channel
	h.subscriptionsMu.RLock()
	// fmt.Println("WS_Handler: Broadcase: Broadcasting to channel:", channel)
	// fmt.Println("WS_Handler: Broadcast: total subscribers:", len(h.subscriptions))
	// Get the list of clients subscribed to the channel
	clients, ok := h.subscriptions[channel]
	h.subscriptionsMu.RUnlock()
	if !ok {
		return
	}

	// Broadcast the message to all clients subscribed to the channel
	for client := range clients {
		// Check if the client has a product filter for the channel
		client.mu.RLock()
		clientProductIDs, hasFilter := client.productFilters[channel]
		client.mu.RUnlock()

		// fmt.Println("WS_Handler: Broadcast: reading product ids:", clientProductIDs)

		// If the client has a product filter, check if the message matches the filter
		if hasFilter && len(clientProductIDs) > 0 {
			// Check if client subscribed to "all" products
			isSubscribedToAll := false
			for _, clientProductID := range clientProductIDs {
				if clientProductID == "all" {
					isSubscribedToAll = true
					break
				}
			}
			
			// If not subscribed to all, check if any of the product IDs match
			if !isSubscribedToAll {
				match := false
				for _, clientProductID := range clientProductIDs {
					if productID == clientProductID {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}
		}

		fmt.Println("WS_Handler: Broadcast: sending message to client:", client.id, "on channel:", channel, "for product:", productID)

		// Send the message to the client
		select {
		case client.send <- message:
		default:
			h.unregister <- client
		}
	}
}

// GetDeltaConnectionStatus gets the connection status of the Delta Exchange client
func (h *WebsocketHandler) GetDeltaConnectionStatus() map[string]interface{} {
	if h.deltaClient != nil {
		return h.deltaClient.GetConnectionStatus()
	}
	return map[string]interface{}{
		"connected": false,
	}
}

// GetStatistics gets statistics about the websocket handler
func (h *WebsocketHandler) GetStatistics() map[string]interface{} {
	// Get the number of active connections
	h.clientsMu.RLock()
	activeConnections := len(h.clients)
	h.clientsMu.RUnlock()

	// Get the number of active subscriptions
	h.subscriptionsMu.RLock()
	activeSubscriptions := 0
	subscriptionsByChannel := make(map[string]int)
	for channel, clients := range h.subscriptions {
		subscriptionsByChannel[channel] = len(clients)
		activeSubscriptions += len(clients)
	}
	h.subscriptionsMu.RUnlock()

	// Get the external sources
	externalSources := make(map[string]bool)
	if h.deltaClient != nil {
		externalSources["delta"] = h.deltaClient.IsConnected()
	}

	// Create the statistics
	stats := map[string]interface{}{
		"active_connections":       activeConnections,
		"active_subscriptions":     activeSubscriptions,
		"messages_sent":            atomic.LoadInt64(&h.messagesSent),
		"messages_received":        atomic.LoadInt64(&h.messagesReceived),
		"subscriptions_by_channel": subscriptionsByChannel,
		"external_sources":         externalSources,
	}

	return stats
}

// Close closes the websocket handler
func (h *WebsocketHandler) Close() {
	h.cancel()
}

// registerDeltaHandler registers a handler for a Delta Exchange channel
func (h *WebsocketHandler) registerDeltaHandler(channel string) {
	h.deltaClient.RegisterHandler(channel, func(message []byte, channelName string, msgProductID string) {
		// Broadcast the message to all clients subscribed to the channel
		h.BroadcastToChannel(channelName, message, msgProductID)
	})
}

// run runs the websocket handler
func (h *WebsocketHandler) run() {
	defer func() {
		// Close all clients
		h.clientsMu.Lock()
		for client := range h.clients {
			client.conn.Close()
		}
		h.clientsMu.Unlock()

		// Close the Delta Exchange client
		if h.deltaClient != nil {
			h.deltaClient.Close()
		}
	}()

	for {
		select {
		case <-h.ctx.Done():
			return
		case client := <-h.register:
			h.clientsMu.Lock()
			h.clients[client] = true
			h.clientsMu.Unlock()
		case client := <-h.unregister:
			h.clientsMu.Lock()
			fmt.Println("WS_Handler: unregistering client")

			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.clientsMu.Unlock()

			// Remove the client from all subscriptions and clean up assets
			client.mu.RLock()
			channelsToUnsubscribe := make([]string, 0, len(client.subscriptions))
			for channel := range client.subscriptions {
				channelsToUnsubscribe = append(channelsToUnsubscribe, channel)
			}
			client.mu.RUnlock()
			
			for _, channel := range channelsToUnsubscribe {
				h.unsubscribeClient(client, channel)
			}
		case message := <-h.broadcast:
			// Broadcast the message to all clients
			h.clientsMu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.clientsMu.RUnlock()
		}
	}
}

// readPump reads messages from the client
func (h *WebsocketHandler) readPump(client *Client) {
	defer func() {
		h.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadLimit(h.config.Websocket.MaxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		// Update the last activity time
		client.lastActivity = time.Now()

		// Increment the messages received counter
		atomic.AddInt64(&h.messagesReceived, 1)

		// Parse the message
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle the message based on the type
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case "subscribe":
				h.handleSubscribe(client, msg)
			case "unsubscribe":
				h.handleUnsubscribe(client, msg)
			case "ping":
				h.handlePing(client)
			default:
				log.Printf("Unknown message type: %s", msgType)
			}
		}
	}
}

// writePump writes messages to the client
func (h *WebsocketHandler) writePump(client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Increment the messages sent counter
			atomic.AddInt64(&h.messagesSent, 1)

			// Add queued messages to the current websocket message
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				msg := <-client.send
				w.Write(msg)
				// Increment the messages sent counter
				atomic.AddInt64(&h.messagesSent, 1)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscribe handles a subscribe message
func (h *WebsocketHandler) handleSubscribe(client *Client, msg map[string]interface{}) {
	fmt.Println("subscribing a message: ", msg)

	var chName string = ""
	// Check if the message has the new format with payload.channels
	if payload, ok := msg["payload"].(map[string]interface{}); ok {
		if channels, ok := payload["channels"].([]interface{}); ok {
			// Process each channel in the payload
			for _, channelObj := range channels {
				if channelMap, ok := channelObj.(map[string]interface{}); ok {
					// Get the channel name
					channelName, ok := channelMap["name"].(string)
					if !ok {
						log.Printf("Channel object does not contain a name")
						continue
					}

					// Get the symbols directly as product IDs
					var productIDs []string
					if symbols, ok := channelMap["symbols"].([]interface{}); ok {
						for _, symbol := range symbols {
							if symbolStr, ok := symbol.(string); ok {
								if symbolStr == "all" {
									// Use all product IDs from config
									productIDs = []string{"all"}
									break
								} else {
									// Add the symbol directly as a product ID
									productIDs = append(productIDs, symbolStr)
								}
							} else if symbolFloat, ok := symbol.(float64); ok {
								// Convert float to string
								productIDs = append(productIDs, fmt.Sprintf("%v", symbolFloat))
							}
						}
					}

					fmt.Println("subscribing to channel: ", channelName)
					fmt.Println("subscribing to product IDs: ", productIDs)
					chName = channelName

					// Check if deltaClient has not registered for this channel then register first.
					if h.deltaClient != nil {
						if channel, ok := msg["type"].(string); ok {
							if channel == "subscribe" {
								// Check if handler is already registered for this channel
								h.handlersMu.RLock()
								isRegistered := h.registeredHandlers[channelName]
								h.handlersMu.RUnlock()
								
								if !isRegistered {
									// Register handler only if not already registered
									h.registerDeltaHandler(channelName)
									h.handlersMu.Lock()
									h.registeredHandlers[channelName] = true
									h.handlersMu.Unlock()
									fmt.Println("WS_handler: Delta: Registered NEW handler for channel: ", channelName)
								} else {
									fmt.Println("WS_handler: Delta: Handler already registered for channel: ", channelName)
								}
								
								fmt.Println("WS_handler: Delta: subscribing to channel: ", channelName)
								h.deltaClient.Subscribe(channelName, productIDs)
							}
						}
					}

					// Subscribe the client to the channel
					h.subscribeClient(client, channelName, productIDs)
				}
			}
		}
	} else {
		log.Printf("Subscribe message does not contain a payload")
		return
	}

	// Get the client's current subscribed assets for this channel
	client.mu.RLock()
	currentAssets := client.productFilters[chName]
	client.mu.RUnlock()

	// Send a subscription confirmation with the actual subscribed assets
	response := map[string]interface{}{
		"type": "subscribed",
		"payload": map[string]interface{}{
			"channels": []map[string]interface{}{
				{
					"name":    chName,
					"symbols": currentAssets,
				},
			},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling subscription confirmation: %v", err)
		return
	}

	fmt.Println("sending subscription confirmation: ", string(data))

	client.send <- data
}

// handleUnsubscribe handles an unsubscribe message
func (h *WebsocketHandler) handleUnsubscribe(client *Client, msg map[string]interface{}) {
	fmt.Println("unsubscribing a message: ", msg)
	
	// Check if the message has the new format with payload.channels
	if payload, ok := msg["payload"].(map[string]interface{}); ok {
		if channels, ok := payload["channels"].([]interface{}); ok {
			// Process each channel in the payload
			for _, channelObj := range channels {
				if channelMap, ok := channelObj.(map[string]interface{}); ok {
					// Get the channel name
					channelName, ok := channelMap["name"].(string)
					if !ok {
						log.Printf("Channel object does not contain a name")
						continue
					}

					// Get the symbols to unsubscribe from
					var assetsToUnsubscribe []string
					if symbols, ok := channelMap["symbols"].([]interface{}); ok {
						for _, symbol := range symbols {
							if symbolStr, ok := symbol.(string); ok {
								if symbolStr == "all" {
									// Unsubscribe from all assets - this means complete channel unsubscription
									fmt.Println("unsubscribing from ALL assets on channel: ", channelName)
									
									// Check if client is actually subscribed to this channel
									client.mu.RLock()
									_, hasSubscription := client.productFilters[channelName]
									client.mu.RUnlock()
									
									if !hasSubscription {
										fmt.Printf("WS_handler: Client not subscribed to channel %s\n", channelName)
										h.sendUnsubscribeError(client, channelName, []string{"all"}, "Not subscribed to this channel")
										return
									}
									
									h.unsubscribeClient(client, channelName)
									// Send unsubscription confirmation
									h.sendUnsubscribeConfirmation(client, channelName, []string{"all"})
									return
								} else {
									assetsToUnsubscribe = append(assetsToUnsubscribe, symbolStr)
								}
							}
						}
					}

					fmt.Println("unsubscribing from channel: ", channelName)
					fmt.Println("unsubscribing from assets: ", assetsToUnsubscribe)

					// Unsubscribe from specific assets
					if len(assetsToUnsubscribe) > 0 {
						success := h.unsubscribeClientFromAssets(client, channelName, assetsToUnsubscribe)
						if success {
							// Send unsubscription confirmation
							h.sendUnsubscribeConfirmation(client, channelName, assetsToUnsubscribe)
						} else {
							// Send error message
							h.sendUnsubscribeError(client, channelName, assetsToUnsubscribe, "Not subscribed to the specified assets")
						}
					}
				}
			}
		}
	} else {
		log.Printf("Unsubscribe message does not contain a payload")
		return
	}
}

// handlePing handles a ping message
func (h *WebsocketHandler) handlePing(client *Client) {
	// Send a pong response
	response := map[string]interface{}{
		"type": "pong",
		"time": time.Now().UnixNano() / int64(time.Millisecond),
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling pong response: %v", err)
		return
	}
	client.send <- data
}

// subscribeClient subscribes a client to a channel
func (h *WebsocketHandler) subscribeClient(client *Client, channel string, productIDs []string) {
	// Add the subscription to the client
	client.mu.Lock()
	client.subscriptions[channel] = true
	
	// Merge new productIDs with existing ones instead of overwriting
	existingProductIDs, exists := client.productFilters[channel]
	if exists {
		// Create a map to avoid duplicates
		productIDMap := make(map[string]bool)
		
		// Add existing product IDs
		for _, existingID := range existingProductIDs {
			productIDMap[existingID] = true
		}
		
		// Add new product IDs
		for _, newID := range productIDs {
			productIDMap[newID] = true
		}
		
		// Convert back to slice
		mergedProductIDs := make([]string, 0, len(productIDMap))
		for productID := range productIDMap {
			mergedProductIDs = append(mergedProductIDs, productID)
		}
		
		client.productFilters[channel] = mergedProductIDs
		fmt.Printf("WS_handler: Merged product filters for channel %s: %v\n", channel, mergedProductIDs)
	} else {
		client.productFilters[channel] = productIDs
		fmt.Printf("WS_handler: Set new product filters for channel %s: %v\n", channel, productIDs)
	}
	
	client.mu.Unlock()

	// Add the client to the subscription
	h.subscriptionsMu.Lock()
	if _, ok := h.subscriptions[channel]; !ok {
		h.subscriptions[channel] = make(map[*Client]bool)
	}
	h.subscriptions[channel][client] = true
	h.subscriptionsMu.Unlock()
}

// getActiveAssetsForChannel returns all assets currently subscribed by active clients for a channel
func (h *WebsocketHandler) getActiveAssetsForChannel(channel string) []string {
	activeAssets := make(map[string]bool)
	
	h.subscriptionsMu.RLock()
	if clients, ok := h.subscriptions[channel]; ok {
		for client := range clients {
			client.mu.RLock()
			if assets, hasFilter := client.productFilters[channel]; hasFilter {
				for _, asset := range assets {
					activeAssets[asset] = true
				}
			}
			client.mu.RUnlock()
		}
	}
	h.subscriptionsMu.RUnlock()
	
	result := make([]string, 0, len(activeAssets))
	for asset := range activeAssets {
		result = append(result, asset)
	}
	return result
}

// sendUnsubscribeConfirmation sends an unsubscribe confirmation to the client
func (h *WebsocketHandler) sendUnsubscribeConfirmation(client *Client, channel string, assets []string) {
	response := map[string]interface{}{
		"type": "unsubscribed",
		"payload": map[string]interface{}{
			"channels": []map[string]interface{}{
				{
					"name":    channel,
					"symbols": assets,
				},
			},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling unsubscription confirmation: %v", err)
		return
	}
	fmt.Println("sending unsubscription confirmation: ", string(data))
	client.send <- data
}

// sendUnsubscribeError sends an unsubscribe error to the client
func (h *WebsocketHandler) sendUnsubscribeError(client *Client, channel string, assets []string, errorMessage string) {
	response := map[string]interface{}{
		"type": "error",
		"error": map[string]interface{}{
			"code":    "UNSUBSCRIBE_FAILED",
			"message": errorMessage,
			"details": map[string]interface{}{
				"channel": channel,
				"symbols": assets,
			},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling unsubscription error: %v", err)
		return
	}
	fmt.Println("sending unsubscription error: ", string(data))
	client.send <- data
}

// sendSubscriptionError sends a subscription error to the client
func (h *WebsocketHandler) sendSubscriptionError(client *Client, channel string, errorMessage string) {
	response := map[string]interface{}{
		"type": "error",
		"error": map[string]interface{}{
			"code":    "SUBSCRIBE_FAILED",
			"message": errorMessage,
			"details": map[string]interface{}{
				"channel": channel,
			},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling subscription error: %v", err)
		return
	}
	fmt.Println("sending subscription error: ", string(data))
	client.send <- data
}

// unsubscribeClientFromAssets unsubscribes a client from specific assets in a channel
// Returns true if successful, false if client wasn't subscribed or assets weren't found
func (h *WebsocketHandler) unsubscribeClientFromAssets(client *Client, channel string, assetsToRemove []string) bool {
	// Get current client assets for this channel
	client.mu.Lock()
	currentAssets, hasSubscription := client.productFilters[channel]
	if !hasSubscription {
		client.mu.Unlock()
		fmt.Printf("WS_handler: Client not subscribed to channel %s\n", channel)
		return false
	}

	// Check if any of the assets to remove actually exist in client's subscription
	removeMap := make(map[string]bool)
	for _, asset := range assetsToRemove {
		removeMap[asset] = true
	}
	
	// Check if client has any of the assets they want to remove
	hasAnyAsset := false
	for _, asset := range currentAssets {
		if removeMap[asset] {
			hasAnyAsset = true
			break
		}
	}
	
	if !hasAnyAsset {
		client.mu.Unlock()
		fmt.Printf("WS_handler: Client not subscribed to any of the assets %v in channel %s\n", assetsToRemove, channel)
		return false
	}

	// Filter out the assets to remove (reuse the removeMap from above)
	var remainingAssets []string
	for _, asset := range currentAssets {
		if !removeMap[asset] {
			remainingAssets = append(remainingAssets, asset)
		}
	}

	fmt.Printf("WS_handler: Client had assets %v, removing %v, remaining %v\n", currentAssets, assetsToRemove, remainingAssets)

	// Update client's product filters
	if len(remainingAssets) == 0 {
		// No assets left, remove client from channel completely
		delete(client.subscriptions, channel)
		delete(client.productFilters, channel)
		client.mu.Unlock()
		
		// Remove client from channel subscription
		h.subscriptionsMu.Lock()
		if clients, ok := h.subscriptions[channel]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.subscriptions, channel)
				h.subscriptionsMu.Unlock()
				
				// No more clients for this channel, unsubscribe completely from Delta
				if h.deltaClient != nil {
					fmt.Printf("WS_handler: Delta: No more clients for channel %s, unsubscribing completely\n", channel)
					h.deltaClient.Unsubscribe(channel)
				}
				return true
			}
		}
		h.subscriptionsMu.Unlock()
		
		fmt.Printf("WS_handler: Client removed from channel %s (no remaining assets)\n", channel)
	} else {
		// Update with remaining assets
		client.productFilters[channel] = remainingAssets
		client.mu.Unlock()
		fmt.Printf("WS_handler: Updated client assets for channel %s: %v\n", channel, remainingAssets)
	}

	// Check if we need to remove specific assets from Delta
	if h.deltaClient != nil && len(assetsToRemove) > 0 {
		// Get assets still needed by all clients
		activeAssets := h.getActiveAssetsForChannel(channel)
		activeAssetsMap := make(map[string]bool)
		for _, asset := range activeAssets {
			activeAssetsMap[asset] = true
		}

		// Find assets that are no longer needed by any client
		var deltaAssetsToRemove []string
		for _, asset := range assetsToRemove {
			// Skip "all" - it's a special case that shouldn't be removed individually
			if asset != "all" && !activeAssetsMap[asset] {
				deltaAssetsToRemove = append(deltaAssetsToRemove, asset)
			}
		}

		// Remove unused assets from Delta subscription
		if len(deltaAssetsToRemove) > 0 {
			fmt.Printf("WS_handler: Delta: Removing unused assets %v from channel %s\n", deltaAssetsToRemove, channel)
			if err := h.deltaClient.UnsubscribeAssets(channel, deltaAssetsToRemove); err != nil {
				fmt.Printf("WS_handler: Delta: Error removing assets: %v\n", err)
			}
		} else {
			fmt.Printf("WS_handler: Delta: Assets %v still needed by other clients, not removing from Delta\n", assetsToRemove)
		}
	}
	
	return true
}

// unsubscribeClient unsubscribes a client from a channel
func (h *WebsocketHandler) unsubscribeClient(client *Client, channel string) {
	// Get the assets this client was subscribed to before removing them
	client.mu.RLock()
	clientAssets, hadSubscription := client.productFilters[channel]
	client.mu.RUnlock()
	
	if !hadSubscription {
		return // Client wasn't subscribed to this channel
	}
	
	// Remove the subscription from the client
	client.mu.Lock()
	delete(client.subscriptions, channel)
	delete(client.productFilters, channel)
	client.mu.Unlock()

	// Remove the client from the subscription
	h.subscriptionsMu.Lock()
	if clients, ok := h.subscriptions[channel]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			// No more clients, unsubscribe completely from Delta
			delete(h.subscriptions, channel)
			h.subscriptionsMu.Unlock()
			
			if h.deltaClient != nil {
				fmt.Printf("WS_handler: Delta: No more clients for channel %s, unsubscribing completely\n", channel)
				h.deltaClient.Unsubscribe(channel)
			}
			return
		}
	}
	h.subscriptionsMu.Unlock()

	// Check if we need to remove specific assets from Delta
	if h.deltaClient != nil && len(clientAssets) > 0 {
		// Get assets still needed by remaining clients
		activeAssets := h.getActiveAssetsForChannel(channel)
		activeAssetsMap := make(map[string]bool)
		for _, asset := range activeAssets {
			activeAssetsMap[asset] = true
		}
		
		// Find assets that are no longer needed
		assetsToRemove := make([]string, 0)
		for _, asset := range clientAssets {
			// Skip "all" - it's a special case that shouldn't be removed individually
			if asset != "all" && !activeAssetsMap[asset] {
				assetsToRemove = append(assetsToRemove, asset)
			}
		}
		
		// Remove unused assets from Delta subscription
		if len(assetsToRemove) > 0 {
			fmt.Printf("WS_handler: Delta: Removing unused assets %v from channel %s\n", assetsToRemove, channel)
			if err := h.deltaClient.UnsubscribeAssets(channel, assetsToRemove); err != nil {
				fmt.Printf("WS_handler: Delta: Error removing assets: %v\n", err)
			}
		}
	}
}
