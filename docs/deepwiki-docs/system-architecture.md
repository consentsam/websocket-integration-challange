# System Architecture

**Relevant source files**
* [README.md](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md)
* [internal/clients/delta_websocket.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go)
* [internal/handlers/websocket_handler.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go)
* [internal/server/server.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go)

This document describes the system architecture of the websocket-service, a real-time data streaming hub that aggregates market data from external sources and distributes it to connected clients through WebSocket and gRPC interfaces. For detailed information about individual components, see [Core Components](#4). For configuration specifics, see [Configuration Guide](#5). For API usage, see [API Reference](#3).

The websocket-service implements a layered architecture that separates external data ingestion, internal processing, and client-facing interfaces to provide scalable real-time data distribution with subscription-based filtering capabilities.

## Architectural Overview

The websocket-service follows a 4-layer architecture pattern that separates concerns between external data sources, internal processing, and client interfaces:

### Four-Layer Architecture Diagram

**Sources:** [README.md#L17-L24](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L17-L24) [internal/handlers/websocket_handler.go#L30-L46](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L30-L46) [internal/clients/delta_websocket.go#L18-L39](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L18-L39) [internal/server/server.go#L16-L22](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L16-L22)

### Layer Responsibilities

| Layer | Components | Responsibilities |
| :--- | :--- | :--- |
| **External Data Sources** | Delta Exchange WebSocket API | Provides real-time market data streams for various trading instruments |
| **Data Ingestion & Processing** | `DeltaWebsocketClient`, `WebsocketHandler` | Connects to external sources, manages client subscriptions, handles message routing and filtering |
| **Service Interface** | `Server` (gRPC) | Provides programmatic access to service statistics, connection status, and control operations |
| **Client Interfaces** | HTTP Server, gRPC Server | Exposes WebSocket endpoints for real-time clients and gRPC endpoints for internal services |

**Sources:** [README.md#L19-L24](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L19-L24) [internal/handlers/websocket_handler.go#L1-L46](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L1-L46)

## Data Flow Architecture

The service implements a subscription-based message routing system where clients receive only the data they have explicitly subscribed to, with optional product-level filtering.

### Message Flow and Subscription Management

**Sources:** [internal/handlers/websocket_handler.go#L136-L184](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L136-L184) [internal/handlers/websocket_handler.go#L407-L492](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L407-L492) [internal/clients/delta_websocket.go#L113-L201](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L113-L201) [internal/server/server.go#L142-L172](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L142-L172)

### Subscription and Message Filtering Logic

The system implements a two-level filtering mechanism:

1.  **Channel-level subscription**: Clients subscribe to specific channels (e.g., `ticker`, `trades`, `l2_orderbook`)
2.  **Product-level filtering**: Within each channel, clients can filter by specific product IDs or symbols

**Sources:** [internal/handlers/websocket_handler.go#L136-L184](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L136-L184) [internal/handlers/websocket_handler.go#L575-L590](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L575-L590)

## Component Interaction Architecture

The service uses a hub-and-spoke pattern where the `WebsocketHandler` acts as the central coordinator between external data sources and connected clients.

### Core Component Relationships

**Sources:** [internal/handlers/websocket_handler.go#L30-L46](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L30-L46) [internal/clients/delta_websocket.go#L18-L39](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L18-L39) [internal/server/server.go#L16-L22](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/server/server.go#L16-L22) [internal/config/config.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go)

### Client Connection Lifecycle Management

The `WebsocketHandler` manages client connections through a channel-based communication pattern:

**Sources:** [internal/handlers/websocket_handler.go#L245-L304](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L245-L304) [internal/handlers/websocket_handler.go#L107-L134](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L107-L134) [internal/handlers/websocket_handler.go#L306-L356](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L306-L356)

## Message Broadcasting Architecture

The system implements a publish-subscribe pattern where the `WebsocketHandler` acts as a message broker between external data sources and connected clients.

### Broadcasting Flow with Product Filtering

**Sources:** [internal/handlers/websocket_handler.go#L237-L243](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L237-L243) [internal/handlers/websocket_handler.go#L136-L184](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L136-L184) [internal/handlers/websocket_handler.go#L358-L405](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L358-L405)

### Concurrent Processing Model

The service uses goroutines extensively to handle concurrent operations:

| Component | Goroutines | Purpose |
| :--- | :--- | :--- |
| `WebsocketHandler` | `run()` | Main event loop for client registration/unregistration |
| `WebsocketHandler` | `readPump()` per client | Read messages from client WebSocket connections |
| `WebsocketHandler` | `writePump()` per client | Write messages to client WebSocket connections |
| `DeltaWebsocketClient` | `readPump()` | Read messages from Delta Exchange WebSocket |
| `DeltaWebsocketClient` | `reconnect()` | Handle automatic reconnection logic |

**Sources:** [internal/handlers/websocket_handler.go#L102-L105](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L102-L105) [internal/handlers/websocket_handler.go#L131-L133](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L131-L133) [internal/clients/delta_websocket.go#L91-L92](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L91-L92) [internal/clients/delta_websocket.go#L292-L315](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L292-L315)

## Error Handling and Resilience Patterns

The service implements several patterns to ensure reliability and fault tolerance:

### Connection Management and Reconnection

**Sources:** [internal/clients/delta_websocket.go#L60-L103](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L60-L103) [internal/clients/delta_websocket.go#L113-L155](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L113-L155) [internal/clients/delta_websocket.go#L292-L315](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L292-L315)

### Thread Safety and Synchronization

The service uses multiple synchronization primitives to ensure thread safety:

| Component | Mutex Type | Protected Resources |
| :--- | :--- | :--- |
| `DeltaWebsocketClient` | `sync.RWMutex mu` | Connection state, error state, counters |
| `DeltaWebsocketClient` | `sync.RWMutex handlersMu` | Message handlers map |
| `DeltaWebsocketClient` | `sync.RWMutex subscriptionsMu` | Subscriptions map |
| `WebsocketHandler` | `sync.RWMutex clientsMu` | Clients map |
| `WebsocketHandler` | `sync.RWMutex subscriptionsMu` | Subscriptions map |
| `Client` | `sync.RWMutex mu` | Client subscriptions and filters |

**Sources:** [internal/clients/delta_websocket.go#L25-L38](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/clients/delta_websocket.go#L25-L38) [internal/handlers/websocket_handler.go#L30-L46](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L30-L46) [internal/handlers/websocket_handler.go#L18-L28](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/handlers/websocket_handler.go#L18-L28)
