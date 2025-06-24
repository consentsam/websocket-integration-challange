# Core Components

**Relevant source files**
* [README.md](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md)

This page provides detailed documentation of the internal packages that implement the core functionality of the websocket service. These components form the foundation of the system's real-time data streaming, connection management, and external integration capabilities.

For information about the overall system architecture and data flow patterns, see [System Architecture](#2). For API interface specifications, see [API Reference](#3). For configuration details, see [Configuration Guide](#5).

## Component Overview

The websocket service is built around four primary internal components that work together to provide real-time data streaming capabilities:

**Component Interaction Flow**

The components interact in a coordinated pattern where the `DeltaWebsocketClient` ingests external data, the `WebsocketHandler` manages client connections and subscriptions, the gRPC server provides internal API access, and the configuration system governs behavior across all components.

Sources: [README.md#L19-L24](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L19-L24)

## WebSocket Handler

The `WebsocketHandler` serves as the central hub for managing client connections, processing subscription requests, and broadcasting messages to connected clients. This component implements the core websocket protocol logic and maintains connection state.

### Key Responsibilities

| Responsibility | Implementation |
| :--- | :--- |
| Connection Management | Handles websocket upgrade, connection lifecycle, and cleanup |
| Subscription Processing | Manages client subscriptions to channels and product IDs |
| Message Broadcasting | Distributes messages from external sources to subscribed clients |
| Protocol Handling | Processes subscribe, unsubscribe, and ping message types |

### Message Processing Architecture

The handler maintains subscription state per connection and filters incoming messages based on client subscriptions and product ID preferences.

Sources: [README.md#L22-L22](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L22-L22) [README.md#L42-L67](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L42-L67)

## Delta Exchange Integration

The `DeltaWebsocketClient` handles the external integration with Delta Exchange's websocket API, providing reliable data ingestion with automatic reconnection capabilities.

### Connection Management

The client maintains a persistent connection to Delta Exchange and implements robust error handling and reconnection logic:

### Supported Data Channels

The client subscribes to multiple Delta Exchange channels and forwards relevant data to the websocket handler:

| Channel | Purpose | Data Format |
| :--- | :--- | :--- |
| `v2/ticker` | Real-time ticker data | Price, volume, change information |
| `v2/trades` | Trade execution data | Individual trade records |
| `v2/l2_orderbook` | Level 2 order book | Bid/ask depth data |
| `v2/candles` | OHLCV candle data | Time-series price data |

### Product Filtering

The client processes product-specific data for supported trading pairs including `BTC_USDT`, `ETH_USDT`, `SOL_USDT`, `AVAX_USDT`, and `MATIC_USDT`.

Sources: [README.md#L21-L21](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L21-L21) [README.md#L28-L28](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L28-L28) [README.md#L34-L34](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L34-L34)

## gRPC Server

The gRPC server component provides programmatic access to websocket service functionality for internal services and monitoring systems.

### Service Interface

### API Capabilities

The gRPC server exposes internal service state and provides administrative capabilities:

* **Statistics Collection**: Provides real-time metrics about connection counts, subscription states, and message throughput
* **Connection Management**: Allows querying active connections and their subscription states
* **Message Broadcasting**: Enables internal services to broadcast messages to connected clients
* **Health Monitoring**: Supports health checks and service discovery

Sources: [README.md#L23-L23](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L23-L23) [README.md#L69-L71](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L69-L71)

## Configuration System

The configuration system provides environment-specific settings management with validation and hot-reloading capabilities.

### Configuration Structure

### Environment-Specific Behavior

| Environment | Configuration Focus | Key Differences |
| :--- | :--- | :--- |
| Local | Development ease | Relaxed security, debug logging, no rate limits |
| Development | Testing stability | Moderate security, detailed logging, basic rate limits |
| Production | Security & performance | Strict security, optimized logging, comprehensive rate limits |

### Configuration Loading Process

The system loads configuration in a hierarchical manner, allowing environment-specific overrides while maintaining default values for common settings.

Sources: [README.md#L73-L79](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L73-L79) [README.md#L122-L128](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L122-L128)