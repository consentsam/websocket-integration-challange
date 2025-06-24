# Configuration System

**Relevant source files**
* [config/development.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml)
* [config/local.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml)
* [config/production.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml)
* [internal/config/config.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go)

## Purpose and Scope

The Configuration System manages all runtime configuration for the websocket service across different deployment environments. It provides a centralized approach to configure service behavior, security settings, external integrations, and operational parameters through YAML files and environment variables.

This documentation covers the configuration structure, loading mechanisms, and environment-specific settings. For detailed configuration options and deployment scenarios, see [Configuration Guide](#5).

## Configuration Architecture

The configuration system uses a layered approach with environment-specific YAML files and a centralized Go struct for type-safe access to configuration values.

### Configuration Structure Overview

Sources: [internal/config/config.go#L8-L55](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L8-L55) [config/local.yaml#L1-L45](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L1-L45) [config/development.yaml#L1-L48](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L1-L48) [config/production.yaml#L1-L53](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L1-L53)

## Configuration Data Model

The configuration system is built around a main `Config` struct with nested configuration sections for different service components.

### Core Configuration Struct

Sources: [internal/config/config.go#L8-L55](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L8-L55)

## Environment-Specific Configuration

The service supports three deployment environments, each with distinct configuration profiles optimized for their use cases.

### Environment Configuration Matrix

| Configuration Section | Local | Development | Production |
| :--- | :--- | :--- | :--- |
| **Service** | | | |
| `log_level` | `debug` | `info` | `info` |
| `http_port` | `8080` | `8080` | `8080` |
| `grpc_port` | `9090` | `9090` | `9090` |
| **Security** | | | |
| `websocket.auth.required` | `false` | `true` | `true` |
| `websocket.check_origin` | `false` | `true` | `true` |
| `security.cors_allowed_origins` | `*` | `dev.cryptovate.com` | `cryptovate.com` |
| `security.rate_limit_enabled` | `false` | `true` | `true` |
| `security.rate_limit_requests` | `100` | `200` | `500` |
| **Performance** | | | |
| `websocket.read_buffer_size` | `1024` | `4096` | `8192` |
| `websocket.write_buffer_size` | `1024` | `4096` | `8192` |
| `websocket.max_message_size` | `4096` | `8192` | `16384` |
| **Delta Integration** | | | |
| `delta.url` | `socket.india.delta.exchange` | `socket.delta.exchange` | `socket.delta.exchange` |
| `delta.reconnect_max` | `5` | `10` | `20` |
| `delta.channels` | `ticker, trades` | `ticker, trades, l2_orderbook` | `ticker, trades, l2_orderbook, candles` |

Sources: [config/local.yaml#L1-L45](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L1-L45) [config/development.yaml#L1-L48](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L1-L48) [config/production.yaml#L1-L53](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L1-L53)

## Configuration Loading Process

The configuration loading follows a default-first approach with environment variable substitution support.

### Configuration Loading Flow

Sources: [internal/config/config.go#L58-L90](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L58-L90)

### Default Configuration Values

The `LoadConfig` function establishes baseline defaults before applying environment-specific overrides:

Sources: [internal/config/config.go#L60-L73](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L60-L73)

## Configuration Sections

### Websocket Configuration

Controls WebSocket connection behavior, buffer management, and authentication settings.

**Key Fields:**

* `read_buffer_size` / `write_buffer_size`: Socket buffer sizes in bytes
* `max_message_size`: Maximum message size limit
* `check_origin`: Origin validation for security
* `auth.required`: Whether authentication is mandatory
* `auth.secret`: Authentication secret (from environment)

Sources: [internal/config/config.go#L27-L36](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L27-L36)

### Security Configuration

Manages CORS policies and rate limiting to protect the service from abuse.

**Key Fields:**

* `cors_enabled`: Enable cross-origin resource sharing
* `cors_allowed_origins`: Comma-separated list of allowed origins
* `rate_limit_enabled`: Enable request rate limiting
* `rate_limit_requests`: Maximum requests per duration
* `rate_limit_duration`: Rate limiting window in seconds

The `GetCORSAllowedOrigins()` method parses the comma-separated origins string into a slice, defaulting to `["*"]` if empty.

Sources: [internal/config/config.go#L39-L45](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L39-L45) [internal/config/config.go#L93-L98](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L93-L98)

### Delta Exchange Configuration

Configures the external Delta Exchange WebSocket integration.

**Key Fields:**

* `enabled`: Whether Delta integration is active
* `url`: WebSocket endpoint URL
* `channels`: Subscribed data channels (ticker, trades, orderbook, candles)
* `product_ids`: Cryptocurrency pairs to monitor
* `reconnect_max`: Maximum reconnection attempts

Sources: [internal/config/config.go#L8-L15](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L8-L15) [internal/config/config.go#L47-L48](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L47-L48)

### Metrics Configuration

Controls metrics exposure for monitoring and observability.

**Key Fields:**

* `enabled`: Whether metrics collection is active
* `endpoint`: HTTP endpoint path for metrics exposure (typically `/metrics`)

Sources: [internal/config/config.go#L51-L54](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L51-L54)

## Environment Variable Integration

The configuration system supports environment variable substitution using the `${VARIABLE_NAME}` syntax in YAML files. This is primarily used for sensitive values like authentication secrets.

**Example Usage:**

```yaml
websocket:
  auth:
    secret: "${WEBSOCKET_AUTH_SECRET}"
````

The `WEBSOCKET_AUTH_SECRET` environment variable is resolved at runtime, allowing secure configuration management without hardcoding secrets in configuration files.

Sources: [config/development.yaml\#L18-L18](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L18-L18) [config/production.yaml\#L18-L18](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L18-L18)
