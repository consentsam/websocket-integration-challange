# Configuration Guide

**Relevant source files**
* [config/development.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml)
* [config/local.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml)
* [config/production.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml)
* [internal/config/config.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go)

This document provides a comprehensive guide to configuring the websocket-service for different deployment environments. It covers the configuration structure, available options, environment-specific settings, and security considerations.

The configuration system uses YAML files for different environments and supports environment variable overrides for sensitive values. For information about the configuration loading implementation details, see [Core Components](#4.4). For deployment-specific configuration considerations, see [Deployment Guide](#7).

## Configuration System Overview

The websocket-service uses a layered configuration approach with three predefined environment configurations: local, development, and production. The configuration is defined by the `Config` struct in [internal/config/config.go#L17-L55](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L17-L55) and loaded through the `LoadConfig` function at [internal/config/config.go#L58-L90](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L58-L90)

### Configuration Loading Flow

### Configuration Structure Hierarchy

Sources: [internal/config/config.go#L8-L55](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L8-L55)

## Environment Configurations

The service supports three predefined environments, each with different security, performance, and operational characteristics.

### Environment Comparison

| Configuration | Local | Development | Production |
| :--- | :--- | :--- | :--- |
| **Purpose** | Local development | Integration testing | Production deployment |
| **Log Level** | `debug` | `info` | `info` |
| **Authentication** | Disabled | Required | Required |
| **CORS Origins** | `*` (all) | `dev.cryptovate.com` | `cryptovate.com` |
| **Rate Limiting** | Disabled | 200 req/60s | 500 req/60s |
| **Buffer Sizes** | 1024 bytes | 4096 bytes | 8192 bytes |
| **Delta Reconnect** | 5 attempts | 10 attempts | 20 attempts |
| **Product Coverage** | Basic (3 pairs) | Extended (5 pairs) | Full (9 pairs) |

### Local Environment Configuration

The local configuration is optimized for development convenience with relaxed security settings.

**Key characteristics:**

* Authentication disabled ([config/local.yaml#L17-L17](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L17-L17))
* CORS allows all origins ([config/local.yaml#L23-L23](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L23-L23))
* Rate limiting disabled ([config/local.yaml#L24-L24](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L24-L24))
* Debug logging enabled ([config/local.yaml#L6-L6](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L6-L6))
* Smaller buffer sizes for faster iteration

**Data sources:**

* Uses Delta Exchange India endpoint ([config/local.yaml#L31-L31](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L31-L31))
* Limited to ticker and trades channels ([config/local.yaml#L33-L34](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L33-L34))
* Covers 3 major cryptocurrency pairs ([config/local.yaml#L36-L38](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L36-L38))

### Development Environment Configuration

The development configuration balances functionality with security for staging environments.

**Key characteristics:**

* Authentication required with environment variable secret ([config/development.yaml#L17-L18](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L17-L18))
* CORS restricted to development domains ([config/development.yaml#L23-L23](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L23-L23))
* Moderate rate limiting (200 requests per 60 seconds) ([config/development.yaml#L25-L26](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L25-L26))
* Standard buffer sizes ([config/development.yaml#L12-L13](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L12-L13))

**Enhanced features:**

* Includes L2 orderbook data ([config/development.yaml#L35-L35](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L35-L35))
* Extended cryptocurrency coverage ([config/development.yaml#L36-L41](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L36-L41))
* Higher reconnection resilience ([config/development.yaml#L42-L42](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L42-L42))

### Production Environment Configuration

The production configuration provides maximum security, performance, and reliability.

**Key characteristics:**

* Strict authentication requirements ([config/production.yaml#L17-L18](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L17-L18))
* CORS limited to production domains only ([config/production.yaml#L23-L23](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L23-L23))
* Higher rate limits for production load ([config/production.yaml#L25-L26](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L25-L26))
* Optimized buffer sizes for performance ([config/production.yaml#L12-L13](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L12-L13))

**Production features:**

* Full channel coverage including candles ([config/production.yaml#L33-L36](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L33-L36))
* Comprehensive cryptocurrency pair support ([config/production.yaml#L37-L46](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L37-L46))
* Maximum reconnection resilience ([config/production.yaml#L47-L47](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L47-L47))

Sources: [config/local.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml) [config/development.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml) [config/production.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml)

## Configuration Options Reference

### Service Configuration

| Option | Type | Description | Default |
| :--- | :--- | :--- | :--- |
| `service_name` | string | Service identifier for logging and metrics | `websocket-service` |
| `environment` | string | Deployment environment (local/development/production) | `local` |
| `log_level` | string | Logging verbosity (debug/info/warn/error) | `info` |
| `http_port` | int | HTTP server port for WebSocket connections | `8083` |
| `grpc_port` | int | gRPC server port for internal API | `9093` |

### WebSocket Configuration

| Option | Type | Description | Impact |
| :--- | :--- | :--- | :--- |
| `read_buffer_size` | int | Input buffer size in bytes | Memory usage, message throughput |
| `write_buffer_size` | int | Output buffer size in bytes | Memory usage, broadcast performance |
| `max_message_size` | int64 | Maximum message size limit | Client message restrictions |
| `check_origin` | bool | Enable WebSocket origin validation | Security, CORS enforcement |
| `auth.required` | bool | Enable authentication for connections | Security level |
| `auth.secret` | string | Secret key for authentication | Token validation |

### Security Configuration

| Option | Type | Description | Security Impact |
| :--- | :--- | :--- | :--- |
| `cors_enabled` | bool | Enable Cross-Origin Resource Sharing | Browser access control |
| `cors_allowed_origins` | string | Comma-separated allowed origins | Domain access restrictions |
| `rate_limit_enabled` | bool | Enable request rate limiting | DoS protection |
| `rate_limit_requests` | int | Maximum requests per duration | Rate limit threshold |
| `rate_limit_duration` | int | Rate limit window in seconds | Rate limit period |

The `GetCORSAllowedOrigins()` method at [internal/config/config.go#L92-L98](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L92-L98) processes the origins string into an array, defaulting to `["*"]` if not specified.

### Delta Exchange Configuration

| Option | Type | Description | Data Impact |
| :--- | :--- | :--- | :--- |
| `enabled` | bool | Enable Delta Exchange integration | Data source availability |
| `url` | string | WebSocket endpoint URL | Connection target |
| `channels` | []string | Subscribed data channels | Data types received |
| `product_ids` | []string | Cryptocurrency pairs to monitor | Market coverage |
| `reconnect_max` | int | Maximum reconnection attempts | Connection resilience |

**Supported channels:**

* `v2/ticker` - Real-time price updates
* `v2/trades` - Trade execution data
* `v2/l2_orderbook` - Order book depth
* `v2/candles` - OHLCV candle data

**Supported product pairs:**

* Major: `BTC_USDT`, `ETH_USDT`, `SOL_USDT`
* Altcoins: `AVAX_USDT`, `MATIC_USDT`, `DOT_USDT`, `LINK_USDT`, `ADA_USDT`, `XRP_USDT`

### Metrics Configuration

| Option | Type | Description | Default |
| :--- | :--- | :--- | :--- |
| `enabled` | bool | Enable Prometheus metrics | `true` |
| `endpoint` | string | Metrics endpoint path | `/metrics` |

Sources: [internal/config/config.go#L8-L55](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L8-L55)

## Environment Variables and Overrides

The configuration system supports environment variable overrides for sensitive values, particularly authentication secrets.

### Required Environment Variables

```shellscript
# Authentication secret for development and production
WEBSOCKET_AUTH_SECRET=your-secret-key-here
````

The secret is referenced in configuration files using `${WEBSOCKET_AUTH_SECRET}` syntax at [config/development.yaml\#L18-L18](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L18-L18) and [config/production.yaml\#L18-L18](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L18-L18)

### Environment Variable Pattern

Sources: [config/development.yaml\#L18-L18](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L18-L18) [config/production.yaml\#L18-L18](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L18-L18)

## Security Configuration Best Practices

### Authentication Configuration

```yaml
websocket:
  auth:
    required: true
    secret: "${WEBSOCKET_AUTH_SECRET}"
```

**Recommendations:**

  * Always enable authentication in non-local environments
  * Use strong, randomly generated secrets (minimum 32 characters)
  * Rotate secrets regularly in production
  * Never commit secrets to version control

### CORS Configuration

```yaml
security:
  cors_enabled: true
  cors_allowed_origins: "[https://cryptovate.com](https://cryptovate.com),[https://api.cryptovate.com](https://api.cryptovate.com)"
```

**Security levels:**

  * **Local**: `"*"` - Allow all origins for development convenience
  * **Development**: Specific development domains only
  * **Production**: Production domains only with HTTPS enforcement

### Rate Limiting Configuration

```yaml
security:
  rate_limit_enabled: true
  rate_limit_requests: 500
  rate_limit_duration: 60
```

**Capacity planning:**

  * **Local**: Disabled for development testing
  * **Development**: Conservative limits (200 req/60s) for testing
  * **Production**: Higher limits (500 req/60s) for real traffic

Sources: [config/local.yaml\#L20-L26](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/local.yaml#L20-L26) [config/development.yaml\#L20-L26](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/development.yaml#L20-L26) [config/production.yaml\#L20-L26](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L20-L26)

## Configuration Validation and Defaults

The system provides sensible defaults through the `LoadConfig` function's default initialization at [internal/config/config.go\#L60-L73](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L60-L73)

### Default Configuration Values

```go
config := &Config{
    ServiceName: serviceName,
    Environment: "local",
    LogLevel:    "info",
    HTTPPort:    8083,
    GRPCPort:    9093,
    Delta: Delta{
        Enabled:      true,
        URL:          "wss://socket.india.delta.exchange",
        Channels:     []string{"v2/ticker"},
        ProductIDs:   []string{"BTCUSD"},
        ReconnectMax: 5,
    },
}
```

These defaults ensure the service can start with minimal configuration while providing a foundation for environment-specific overrides.

Sources: [internal/config/config.go\#L60-L73](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/internal/config/config.go#L60-L73)