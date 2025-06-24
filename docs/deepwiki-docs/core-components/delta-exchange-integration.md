# Delta Exchange Integration

**Relevant source files**
* [README.md](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md)
* [internal/clients/delta_websocket.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go)

The Delta Exchange Integration component provides the external data source client for connecting to and consuming real-time market data from Delta Exchange's WebSocket API. This component is responsible for establishing and maintaining persistent connections, handling message subscriptions, and forwarding market data to the internal WebSocket handler for client distribution.

For information about managing client connections and subscriptions, see [WebSocket Handler](#4.1). For gRPC service implementation details, see [gRPC Server](#4.3).

## Purpose and Responsibilities

The `DeltaWebsocketClient` serves as the primary interface between the websocket-service and Delta Exchange's real-time market data streams. It handles:

* Establishing and maintaining WebSocket connections to Delta Exchange
* Automatic reconnection with configurable retry limits
* Channel subscription and message filtering management
* Thread-safe message processing and handler dispatch
* Connection status monitoring and error tracking
* Protocol translation between Delta Exchange API format and internal message format

## Core Architecture

### Client Structure and Dependencies

Sources: [internal/clients/delta_websocket.go#L1-L376](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L1-L376)

### Connection Lifecycle Management

Sources: [internal/clients/delta_websocket.go#L60-L103](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L60-L103) [internal/clients/delta_websocket.go#L292-L315](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L292-L315)

## Message Processing Architecture

### Subscription Protocol Implementation

The client implements Delta Exchange's subscription protocol with payload-based channel management:

| Message Type | Purpose | Format |
| :--- | :--- | :--- |
| `subscribe` | Subscribe to channels | `{"type": "subscribe", "payload": {"channels": [{"name": "ticker", "symbols": ["BTC_USDT"]}]}}` |
| `unsubscribe` | Unsubscribe from channels | `{"type": "unsubscribe", "payload": {"channels": [{"name": "ticker"}]}}` |
| `ticker` | Market ticker data | `{"type": "ticker", "symbol": "BTC_USDT", ...}` |
| `trades` | Trade execution data | `{"type": "trades", "symbol": "ETH_USDT", ...}` |

Sources: [internal/clients/delta_websocket.go#L204-L253](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L204-L253) [internal/clients/delta_websocket.go#L255-L290](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L255-L290)

### Message Handler Dispatch System

Sources: [internal/clients/delta_websocket.go#L112-L202](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L112-L202) [internal/clients/delta_websocket.go#L106-L110](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L106-L110)

## Thread Safety and Concurrency

### Mutex Protection Strategy

The `DeltaWebsocketClient` uses multiple mutexes to ensure thread-safe operations:

| Mutex | Protected Resources | Usage Pattern |
| :--- | :--- | :--- |
| `mu` | Connection state, error tracking, reconnect counters | `sync.RWMutex` for read-heavy connection status checks |
| `handlersMu` | Message handler registry | `sync.RWMutex` for handler registration and lookup |
| `subscriptionsMu` | Active subscription tracking | `sync.RWMutex` for subscription management |

**Critical Sections:**

* Connection establishment: [internal/clients/delta_websocket.go#L64-L89](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L64-L89)
* Message handler dispatch: [internal/clients/delta_websocket.go#L188-L200](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L188-L200)
* Subscription state updates: [internal/clients/delta_websocket.go#L247-L250](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L247-L250)

Sources: [internal/clients/delta_websocket.go#L19-L39](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L19-L39) [internal/clients/delta_websocket.go#L64-L89](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L64-L89)

### Reconnection Logic Implementation

Sources: [internal/clients/delta_websocket.go#L292-L315](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L292-L315) [internal/clients/delta_websocket.go#L125-L143](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L125-L143)

## Configuration Integration

### Delta Configuration Structure

The client integrates with the configuration system through the `config.Delta` structure:

```go
// Configuration fields used by DeltaWebsocketClient
type Delta struct {
    URL          string   // WebSocket endpoint URL
    Channels     []string // Default channels to subscribe
    ProductIDs   []string // Default product identifiers  
    ReconnectMax  int      // Maximum reconnection attempts
}
````

**Default Values:**

  * `URL`: `"wss://socket.delta.exchange"`
  * `ReconnectMax`: Configurable per environment (5-20 attempts)
  * `ReconnectDelay`: Fixed at 5 seconds

Sources: [internal/clients/delta\_websocket.go\#L42-L58](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L42-L58) [internal/clients/delta\_websocket.go\#L52-L53](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L52-L53)

## Error Handling and Monitoring

### Connection Status Tracking

The client provides comprehensive status monitoring through the `GetConnectionStatus()` method:

| Status Field | Type | Description |
| :--- | :--- | :--- |
| `connected` | `bool` | Current connection state |
| `connected_at` | `time.Time` | Timestamp of last successful connection |
| `reconnect_count` | `int` | Current reconnection attempt count |
| `reconnect_max` | `int` | Maximum allowed reconnection attempts |
| `last_error` | `string` | Most recent error message |
| `last_error_at` | `time.Time` | Timestamp of last error |
| `subscribed_channels` | `[]string` | Active channel subscriptions |

Sources: [internal/clients/delta\_websocket.go\#L345-L361](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L345-L361) [internal/clients/delta\_websocket.go\#L363-L375](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L363-L375)

### Error Recovery Patterns

Sources: [internal/clients/delta\_websocket.go\#L148-L155](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L148-L155) [internal/clients/delta\_websocket.go\#L159-L162](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L159-L162) [internal/clients/delta\_websocket.go\#L301-L304](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L301-L304)

## API Methods and Usage

### Core Public Methods

| Method | Purpose | Thread Safe | Return Type |
| :--- | :--- | :--- | :--- |
| `NewDeltaWebsocketClient()` | Constructor with configuration | N/A | `*DeltaWebsocketClient` |
| `Connect()` | Establish WebSocket connection | Yes | `error` |
| `RegisterHandler()` | Register message handler for channel | Yes | `void` |
| `Subscribe()` | Subscribe to channel with product filters | Yes | `error` |
| `Unsubscribe()` | Unsubscribe from channel | Yes | `error` |
| `IsConnected()` | Check connection status | Yes | `bool` |
| `GetConnectionStatus()` | Get detailed status information | Yes | `map[string]interface{}` |
| `Close()` | Gracefully close connection | Yes | `error` |

Sources: [internal/clients/delta\_websocket.go\#L41-L58](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L41-L58) [internal/clients/delta\_websocket.go\#L60-L103](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L60-L103) [internal/clients/delta\_websocket.go\#L204-L253](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L204-L253)

### Integration with WebSocket Handler

The Delta Exchange client integrates with the internal WebSocket handler through the `MessageHandler` function type:

```go
type MessageHandler func(message []byte, productId string)
```

**Handler Registration Pattern:**

1.  WebSocket handler registers a handler function using `RegisterHandler()`
2.  Delta client calls the handler for each received message matching the channel
3.  Handler function receives raw message bytes and extracted product ID
4.  WebSocket handler processes and distributes to subscribed clients

Sources: [internal/clients/delta\_websocket.go\#L15-L16](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L15-L16) [internal/clients/delta\_websocket.go\#L106-L110](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L106-L110) [internal/clients/delta\_websocket.go\#L196-L200](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L196-L200)