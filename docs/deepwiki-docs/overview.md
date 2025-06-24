<markdownoutput><![CDATA[
# Overview

<details>
<summary>Relevant source files</summary>

- [README.md](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md)  
- [main.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go)

</details>

This document provides an overview of the **websocket-service**, a real-time data-streaming microservice that forms part of the Cryptovate platform. The service acts as a data-aggregation and distribution hub, connecting to external market-data sources and exposing WebSocket and gRPC APIs for real-time consumption.

For detailed information about individual components, see **[Core Components](#core-components)**.  
For API specifications and endpoint documentation, see **API Reference**.  
For deployment and operational guidance, see **Deployment Guide**.

---

## Purpose and Scope

The websocket-service is designed to:

- **Aggregate real-time market data** from external sources like Delta Exchange  
- **Manage client connections** with subscription-based data filtering  
- **Provide dual API interfaces** via WebSocket for real-time clients and gRPC for internal services  
- **Handle connection resilience** with automatic reconnection and error recovery  
- **Support horizontal scaling** through a stateless design and external configuration  

The service operates as a stateless microservice that can be deployed behind load-balancers and integrated with API-gateways for public access.

---

## System Architecture

The websocket-service implements a four-layer architecture that separates concerns between external data ingestion, internal processing, and client-facing interfaces.

### High-Level Architecture

*(Mermaid flow-chart omitted for brevity.)*

---

## Core Components

The service consists of several key components that work together to provide real-time data-streaming capabilities.

| Component             | File path               | Purpose                                                         |
|-----------------------|-------------------------|-----------------------------------------------------------------|
| WebsocketHandler      | internal/handlers/      | Manages client connections, subscriptions, and broadcasting     |
| DeltaWebsocketClient  | internal/clients/       | Handles the connection to Delta Exchange WebSocket API          |
| WebsocketServer       | internal/server/        | Implements the gRPC service interface for internal API access   |
| Config                | internal/config/        | Loads and validates environment-specific configuration          |
| main()                | main.go                 | Application bootstrap and server lifecycle management           |

### Data-Flow Architecture

*(Mermaid sequence diagram omitted for brevity.)*

---

## Integration Points

### WebSocket API

- **Endpoint:** `/ws` on HTTP port 8080  
- **Protocol:** WebSocket with JSON message format  
- **Authentication:** configurable via `WEBSOCKET_AUTH_SECRET`  
- **Message types:** subscribe, unsubscribe, ping  

### gRPC API

- **Port:** 9090 (configurable)  
- **Service:** websocketv1.WebsocketServiceServer  
- **Reflection:** enabled for service discovery  
- **Purpose:** internal service-to-service communication  

### Operational Endpoints

- **Health-check:** `/health` returns HTTP 200 with “OK”  
- **Metrics:** `/metrics` (if enabled) returns JSON statistics  
- **Statistics exposed:** active connections, subscriptions, message counts  

### Configuration Management

The service loads environment-specific YAML configuration files to populate ports, credentials, and feature flags at startup. A dedicated configuration loader validates and exposes these settings to each component.

]]></markdownoutput>
