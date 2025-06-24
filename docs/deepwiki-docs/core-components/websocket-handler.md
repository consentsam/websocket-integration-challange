# WebSocket Handler

**Relevant source files**
* [README.md](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md)
* [internal/handlers/websocket_handler.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go)

The WebSocket Handler is the core component responsible for managing client WebSocket connections, handling channel subscriptions, and broadcasting real-time market data to connected clients. It serves as the central hub that coordinates between external data sources (like Delta Exchange) and client applications, implementing subscription-based filtering and connection lifecycle management.

For information about the external data source integration, see [Delta Exchange Integration](#4.2). For the gRPC API that provides programmatic access to handler statistics, see [gRPC Server](#4.3).

## Architecture Overview

The WebSocket Handler implements a subscription-based broadcasting system with the following key components:

### Core Components Architecture

Sources: [internal/handlers/websocket_handler.go#L18-L46](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L18-L46)

### Message Flow and Client Lifecycle

Sources: [internal/handlers/websocket_handler.go#L108-L134](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L108-L134) [internal/handlers/websocket_handler.go#L246-L304](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L246-L304) [internal/handlers/websocket_handler.go#L136-L184](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L136-L184)

## Connection Management

### Client Registration Process

The handler manages client connections through a registration system implemented in [internal/handlers/websocket_handler.go#L108-L134](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L108-L134):

| Component | Purpose | Key Functions |
| :--- | :--- | :--- |
| `HandleWebsocket` | HTTP upgrade and client creation | Upgrades HTTP to WebSocket, creates `Client` struct |
| `register` channel | Client registration queue | Receives new clients for registration |
| `unregister` channel | Client cleanup queue | Receives clients for disconnection cleanup |
| `clients` map | Active client tracking | Maps `*Client` to boolean for O(1) lookup |

### Client Structure and Properties

Sources: [internal/handlers/websocket_handler.go#L18-L28](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L18-L28)

### Connection Lifecycle Management

The handler implements concurrent read and write operations for each client:

| Operation | Implementation | Purpose |
| :--- | :--- | :--- |
| **Read Pump** | `readPump(client)` | Handles incoming client messages, processes subscriptions |
| **Write Pump** | `writePump(client)` | Sends messages to client, handles ping/pong keepalive |
| **Cleanup** | `unregister` channel | Removes client from all subscriptions and closes connections |

Sources: [internal/handlers/websocket_handler.go#L306-L356](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L306-L356) [internal/handlers/websocket_handler.go#L358-L405](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L358-L405)

## Subscription Management

### Channel Subscription System

The handler implements a two-level subscription system with channel-based grouping and product-level filtering:

Sources: [internal/handlers/websocket_handler.go#L38-L39](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L38-L39) [internal/handlers/websocket_handler.go#L575-L590](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L575-L590)

### Message Format and Handling

The handler supports three message types with specific JSON formats:

| Message Type | Handler Function | Purpose |
| :--- | :--- | :--- |
| `subscribe` | `handleSubscribe()` | Add client to channel with product filters |
| `unsubscribe` | `handleUnsubscribe()` | Remove client from channel |
| `ping` | `handlePing()` | Keepalive mechanism |

#### Subscribe Message Format

```json
{
  "type": "subscribe",
  "payload": {
    "channels": [
      {
        "name": "v2/ticker",
        "symbols": ["BTC_USDT", "ETH_USDT", "all"]
      }
    ]
  }
}
````

Sources: [internal/handlers/websocket\_handler.go\#L407-L492](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L407-L492)

## Broadcasting and Filtering

### Message Broadcasting Architecture

The core broadcasting logic implements product-level filtering to ensure clients only receive relevant data:

Sources: [internal/handlers/websocket\_handler.go\#L237-L243](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L237-L243) [internal/handlers/websocket\_handler.go\#L136-L184](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L136-L184) [internal/handlers/websocket\_handler.go\#L358-L405](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L358-L405)

### Product Filtering Logic

The filtering system supports both specific product IDs and wildcard matching:

| Filter Type | Example | Behavior |
| :--- | :--- | :--- |
| Specific Products | `["BTC_USDT", "ETH_USDT"]` | Only messages for these products |
| Wildcard | `["all"]` | All messages for the channel |
| Mixed | `["BTC_USDT", "all"]` | All messages (wildcard takes precedence) |

Sources: [internal/handlers/websocket\_handler.go\#L158-L173](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L158-L173)

## Integration Points

### Delta Exchange Integration

The handler coordinates with the Delta Exchange client for external data:

Sources: [internal/handlers/websocket\_handler.go\#L237-L243](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L237-L243) [internal/handlers/websocket\_handler.go\#L449-L458](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L449-L458)

### Statistics and Monitoring

The handler provides comprehensive statistics through the `GetStatistics()` method:

| Metric | Source | Purpose |
| :--- | :--- | :--- |
| `active_connections` | `len(h.clients)` | Number of connected clients |
| `active_subscriptions` | Sum of channel subscriptions | Total subscription count |
| `messages_sent` | `atomic.LoadInt64(&h.messagesSent)` | Outbound message counter |
| `messages_received` | `atomic.LoadInt64(&h.messagesReceived)` | Inbound message counter |
| `subscriptions_by_channel` | Per-channel client counts | Channel-specific metrics |
| `external_sources` | Delta client connection status | External connectivity health |

Sources: [internal/handlers/websocket\_handler.go\#L196-L230](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L196-L230)

## Configuration and Initialization

### Handler Configuration

The handler is configured through the `Config` struct with the following key settings:

| Configuration | Purpose | Default Behavior |
| :--- | :--- | :--- |
| `ReadBufferSize` | WebSocket read buffer | From config |
| `WriteBufferSize` | WebSocket write buffer | From config |
| `CheckOrigin` | CORS validation | Configurable per environment |
| `MaxMessageSize` | Message size limit | Applied to all clients |

### Initialization Process

Sources: [internal/handlers/websocket\_handler.go\#L48-L106](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L48-L106)