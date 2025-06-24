# API Reference

**Relevant source files**
* [README.md](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md)
* [protos/websocket/v1/api.proto](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/protos/websocket/v1/api.proto)

This page provides a complete reference for all public interfaces exposed by the websocket-service. The service provides three distinct API layers: WebSocket for real-time client connections, gRPC for internal service integration, and HTTP for health monitoring and WebSocket upgrades.

For information about the internal component implementations, see [Core Components](#4). For deployment-specific configuration of these APIs, see [Configuration Guide](#5).

## API Layer Overview

The websocket-service exposes multiple protocol interfaces to serve different client types and use cases:

**API Protocol Distribution Diagram**

Sources: [README.md#L36-L71](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L36-L71) [protos/websocket/v1/api.proto#L10-L29](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/protos/websocket/v1/api.proto#L10-L29)

## WebSocket API

### Connection Endpoint

Clients connect to the WebSocket endpoint at `/ws` to establish real-time data streams. The connection supports JSON-based message exchange for subscription management.

**Connection URL**: `ws://hostname:8080/ws` or `wss://hostname:8080/ws`

### Message Types

The WebSocket API supports three primary message types for client interaction:

**WebSocket Message Flow Diagram**

#### Subscribe Message

Subscribe to specific channels with optional product ID filtering:

```json
{
  "type": "subscribe",
  "channel": "v2/ticker",
  "product_ids": [27, 28, 29]
}
````

**Supported Channels:**

  * `v2/ticker` - Real-time price ticker data
  * `v2/trades` - Trade execution data
  * `v2/l2_orderbook` - Level 2 order book updates
  * `v2/candles` - Candlestick/OHLCV data

**Product ID Filtering:**
Product IDs correspond to trading pairs on Delta Exchange:

  * `27` - BTC\_USDT
  * `28` - ETH\_USDT
  * `29` - SOL\_USDT
  * Additional product IDs as supported by Delta Exchange

#### Unsubscribe Message

Remove subscription from a specific channel:

```json
{
  "type": "unsubscribe", 
  "channel": "v2/ticker"
}
```

#### Ping Message

Maintain connection liveness:

```json
{
  "type": "ping"
}
```

**Response:**

```json
{
  "type": "pong"
}
```

Sources: [README.md\#L44-L67](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L44-L67)

## gRPC API

The `WebsocketService` provides programmatic access to websocket functionality for internal services. All methods use Protocol Buffer message definitions.

### Service Definition

**gRPC Service Method Mapping**

### Core Methods

#### Subscribe

Creates a new subscription to a channel with optional product filtering.

**Request:** `SubscribeRequest`

  * `channel` (string) - Channel name to subscribe to
  * `product_ids` (repeated string) - Optional product ID filters

**Response:** `SubscribeResponse`

  * `subscription_id` (string) - Unique subscription identifier
  * `channel` (string) - Subscribed channel name
  * `product_ids` (repeated string) - Applied product filters
  * `created_at` (timestamp) - Subscription creation time

#### GetStatistics

Retrieves operational statistics about the websocket service.

**Request:** `google.protobuf.Empty`

**Response:** `StatisticsResponse`

  * `active_connections` (int32) - Current WebSocket connections
  * `active_subscriptions` (int32) - Total active subscriptions
  * `messages_sent` (int64) - Cumulative messages sent to clients
  * `messages_received` (int64) - Cumulative messages from external sources
  * `subscriptions_by_channel` (map\<string, int32\>) - Subscription count per channel
  * `external_sources` (map\<string, bool\>) - External source connection status

#### GetConnectionStatus

Returns the connection status to external data sources.

**Request:** `google.protobuf.Empty`

**Response:** `ConnectionStatusResponse`

  * `connections` (map\<string, ConnectionStatus\>) - Status by source name

**ConnectionStatus Fields:**

  * `connected` (bool) - Current connection state
  * `connected_at` (timestamp) - Connection establishment time
  * `reconnect_attempts` (int32) - Failed reconnection count
  * `last_error` (string) - Most recent error message
  * `last_error_at` (timestamp) - Error occurrence time

### Administrative Methods

| Method | Purpose | Authentication Required |
| :--- | :--- | :--- |
| `Subscribe` | Create channel subscriptions | Internal only |
| `Unsubscribe` | Remove subscriptions | Internal only |
| `GetSubscriptionStatus` | Query subscription details | Internal only |
| `Broadcast` | Send messages to subscribers | Internal only |
| `GetConnectionStatus` | Monitor external connections | Internal only |
| `GetStatistics` | Retrieve service metrics | Internal only |

Sources: [protos/websocket/v1/api.proto\#L10-L149](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/protos/websocket/v1/api.proto#L10-L149)

## HTTP API

The HTTP server provides operational endpoints for health monitoring and metrics collection.

### Endpoints

**HTTP Endpoint Architecture**

#### WebSocket Upgrade Endpoint

**Path:** `/ws`
**Method:** GET with `Upgrade: websocket` header
**Purpose:** Establish WebSocket connections for real-time data streaming

**Headers Required:**

  * `Connection: Upgrade`
  * `Upgrade: websocket`
  * `Sec-WebSocket-Key: <base64-key>`
  * `Sec-WebSocket-Version: 13`

#### Health Check Endpoint

**Path:** `/health`
**Method:** GET
**Purpose:** Service health verification for load balancers and monitoring

**Response Format:**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "dependencies": {
    "delta_exchange": "connected",
    "grpc_server": "running"
  }  
}
```

#### Metrics Endpoint

**Path:** `/metrics`
**Method:** GET
**Content-Type:** `text/plain`
**Purpose:** Prometheus-compatible metrics for observability

**Sample Metrics:**

```
# HELP websocket_active_connections Current number of WebSocket connections
# TYPE websocket_active_connections gauge
websocket_active_connections 42

# HELP websocket_messages_sent_total Total messages sent to clients  
# TYPE websocket_messages_sent_total counter
websocket_messages_sent_total 15847

# HELP websocket_subscription_count Subscriptions by channel
# TYPE websocket_subscription_count gauge
websocket_subscription_count{channel="v2/ticker"} 15
websocket_subscription_count{channel="v2/trades"} 8
```

Sources: [README.md\#L36-L71](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L36-L71) [README.md\#L125-L127](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L125-L127)

## Data Formats

### WebSocket Message Structure

All WebSocket messages follow a consistent JSON structure with message type discrimination:

```json
{
  "type": "message_type",
  "channel": "channel_name", 
  "data": { ... },
  "timestamp": "2024-01-01T00:00:00Z",
  "product_id": "BTC_USDT"
}
```

### Product ID Mapping

The service maps string product identifiers to Delta Exchange numeric IDs:

| Product String | Delta Exchange ID | Description |
| :--- | :--- | :--- |
| `BTC_USDT` | 27 | Bitcoin/Tether |
| `ETH_USDT` | 28 | Ethereum/Tether |
| `SOL_USDT` | 29 | Solana/Tether |
| `AVAX_USDT` | 139 | Avalanche/Tether |
| `MATIC_USDT` | 140 | Polygon/Tether |

### Channel Data Formats

#### Ticker Data (`v2/ticker`)

```json
{
  "type": "ticker",
  "channel": "v2/ticker", 
  "data": {
    "symbol": "BTC_USDT",
    "price": "42000.50",
    "change_24h": "2.5",
    "volume_24h": "1250000",
    "high_24h": "42500.00",
    "low_24h": "41000.00"
  }
}
```

#### Trade Data (`v2/trades`)

```json
{
  "type": "trade",
  "channel": "v2/trades",
  "data": {
    "symbol": "BTC_USDT", 
    "price": "42000.50",
    "size": "0.1",
    "side": "buy",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

Sources: [README.md\#L44-L49](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L44-L49)

## Authentication and Security

### WebSocket Authentication

WebSocket connections require authentication tokens passed via query parameters or headers:

**Query Parameter:** `?token=<jwt_token>`
**Header:** `Authorization: Bearer <jwt_token>`

### gRPC Authentication

gRPC calls require internal service authentication through mutual TLS or service tokens configured per environment.

### Rate Limiting

Rate limits apply per client connection:

| Environment | Rate Limit | Window |
| :--- | :--- | :--- |
| Development | Unlimited | - |
| Production | 500 requests | 60 seconds |

### CORS Policy

WebSocket connections restricted to allowed origins:

  * `https://*.cryptovate.com`
  * `https://localhost:*` (development only)

Sources: [README.md\#L124-L128](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L124-L128)