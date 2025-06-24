# gRPC Server

**Relevant source files**
* [README.md](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md)
* [internal/server/server.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go)
* [protos/websocket/v1/api.proto](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/protos/websocket/v1/api.proto)

This document covers the internal gRPC server implementation that provides programmatic access to websocket service functionality for other internal services. The gRPC server exposes methods for subscription management, connection monitoring, and statistics collection.

For client-facing WebSocket connection management, see [WebSocket Handler](#4.1). For external data source integration, see [Delta Exchange Integration](#4.2).

## Purpose and Architecture

The gRPC server acts as an internal service API layer, allowing other microservices within the Cryptovate platform to programmatically interact with the websocket service without establishing direct WebSocket connections. It provides a typed, efficient interface for managing subscriptions, monitoring connection health, and retrieving operational statistics.

### Component Architecture

**gRPC Server Integration Architecture**

Sources: [internal/server/server.go#L16-L31](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L16-L31) [protos/websocket/v1/api.proto#L10-L29](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/protos/websocket/v1/api.proto#L10-L29)

## Server Implementation

The `Server` struct implements the `UnimplementedWebsocketServiceServer` interface generated from the Protocol Buffers definition. It maintains references to the application context, configuration, and the core `WebsocketHandler` for delegating operations.

### Core Server Structure

| Component | Type | Purpose |
| :--- | :--- | :--- |
| `ctx` | `context.Context` | Application context for request handling |
| `config` | `*config.Config` | Service configuration access |
| `websocketHandler` | `*handlers.WebsocketHandler` | Core business logic delegation |

The server is instantiated through `NewServer()` function which injects dependencies:

**Server Initialization Flow**

Sources: [internal/server/server.go#L24-L31](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L24-L31)

## API Methods Implementation

The gRPC server implements six primary methods defined in the `WebsocketService` interface:

### Subscription Management Methods

#### Subscribe Method

Handles subscription requests by generating unique subscription IDs and returning subscription metadata. The method signature follows the pattern:

* **Input**: `SubscribeRequest` containing `channel` and `product_ids`
* **Output**: `SubscribeResponse` with generated `subscription_id`, timestamp, and request details
* **Implementation**: [internal/server/server.go#L33-L47](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L33-L47)

#### Unsubscribe Method

Processes unsubscription requests using subscription IDs:

* **Input**: `UnsubscribeRequest` with `subscription_id`
* **Output**: `google.protobuf.Empty`
* **Implementation**: [internal/server/server.go#L49-L55](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L49-L55)

### Status and Monitoring Methods

#### GetSubscriptionStatus Method

Retrieves subscription information by querying the `websocketHandler.GetStatistics()`:

* **Data Source**: `stats["subscriptions_by_channel"]`
* **Processing**: Converts channel statistics to `Subscription` proto messages
* **Implementation**: [internal/server/server.go#L57-L82](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L57-L82)

#### GetConnectionStatus Method

Monitors external data source connectivity, specifically Delta Exchange:

* **Condition Check**: `config.Delta.Enabled`
* **Status Fields**: `connected`, `connected_at`, `reconnect_count`, `last_error`, `last_error_at`
* **Implementation**: [internal/server/server.go#L84-L123](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L84-L123)

#### GetStatistics Method

Aggregates comprehensive service metrics:

**Statistics Data Flow**

Sources: [internal/server/server.go#L142-L172](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L142-L172)

#### Broadcast Method

Enables programmatic message broadcasting to channel subscribers:

* **Message Processing**: Converts `int64` product IDs to strings for handler compatibility
* **Delegation**: Calls `websocketHandler.BroadcastToChannel()`
* **Implementation**: [internal/server/server.go#L125-L140](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L125-L140)

## Protocol Buffers Definition

The service interface is formally defined in the Protocol Buffers schema:

### Service Definition Structure

| Method | Request Type | Response Type | Purpose |
| :--- | :--- | :--- | :--- |
| `Subscribe` | `SubscribeRequest` | `SubscribeResponse` | Create channel subscription |
| `Unsubscribe` | `UnsubscribeRequest` | `Empty` | Remove subscription |
| `GetSubscriptionStatus` | `GetSubscriptionStatusRequest` | `GetSubscriptionStatusResponse` | Query subscription state |
| `GetConnectionStatus` | `Empty` | `ConnectionStatusResponse` | Monitor external connections |
| `Broadcast` | `BroadcastRequest` | `Empty` | Send message to subscribers |
| `GetStatistics` | `Empty` | `StatisticsResponse` | Retrieve service metrics |

### Key Message Types

The proto definition includes specialized message types for different operations:

* **Subscription Messages**: `SubscribeRequest`, `SubscribeResponse`, `Subscription`
* **Status Messages**: `ConnectionStatus`, `ConnectionStatusResponse`
* **Operational Messages**: `BroadcastRequest`, `StatisticsResponse`

Sources: [protos/websocket/v1/api.proto#L10-L149](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/protos/websocket/v1/api.proto#L10-L149)

## Integration Patterns

The gRPC server serves as a bridge between internal services and the websocket functionality:

**gRPC Integration Use Cases**

The server enables programmatic access patterns while maintaining separation between external client connections (handled by WebSocket endpoints) and internal service operations (handled by gRPC methods).

Sources: [internal/server/server.go#L1-L173](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L1-L173) [protos/websocket/v1/api.proto#L1-L149](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/protos/websocket/v1/api.proto#L1-L149)