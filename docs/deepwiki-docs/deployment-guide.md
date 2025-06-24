# Deployment Guide

**Relevant source files**
* [Dockerfile](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Dockerfile)
* [README.md](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md)
* [config/production.yaml](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml)

This document provides comprehensive instructions for deploying the websocket-service to production environments. It covers containerization, orchestration, configuration management, security considerations, and operational best practices.

For development setup and local testing, see [Development Guide](#6). For configuration options and environment-specific settings, see [Configuration Guide](#5).

## Prerequisites and Requirements

The websocket-service requires the following infrastructure components and dependencies for production deployment:

### System Requirements

* Container runtime (Docker 20.10+ or compatible)
* Kubernetes cluster (1.24+ recommended) for orchestrated deployments
* Load balancer with WebSocket support
* TLS certificate management
* Secret management system

### Network Requirements

* Outbound connectivity to Delta Exchange WebSocket API (`wss://socket.delta.exchange`)
* Inbound access on HTTP port 8080 and gRPC port 9090
* API Gateway integration for public WebSocket endpoint routing

### Dependencies

* External data source: Delta Exchange WebSocket API
* Configuration files: [`config/production.yaml`](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/`config/production.yaml`)
* Authentication secrets: `WEBSOCKET_AUTH_SECRET` environment variable

Sources: [README.md#L83-L129](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L83-L129) [Dockerfile#L34-L35](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Dockerfile#L34-L35) [config/production.yaml#L17-L18](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L17-L18)

## Deployment Architecture

The websocket-service deployment follows a multi-tier architecture with external integrations and internal service communication:

### Production Deployment Flow

Sources: [README.md#L17-L24](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L17-L24) [README.md#L118-L121](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L118-L121) [config/production.yaml#L22-L26](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L22-L26)

## Container Deployment

### Docker Image Build

The service uses a multi-stage Docker build process optimized for production:

### Build Process Components

**Docker Commands**

| Command | Purpose |
| :--- | :--- |
| `make docker-build` | Build container image with tag `cryptovate/websocket-service:latest` |
| `make docker-run` | Run container with ports 8080 and 9090 exposed |

Sources: [Makefile#L46-L53](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L46-L53)

### Container Configuration

The container exposes two ports as defined in [Dockerfile#L34-L35](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Dockerfile#L34-L35):

* Port 8080: HTTP server for WebSocket connections and health endpoints
* Port 9090: gRPC server for internal service communication

The container requires configuration files mounted from [`config/production.yaml`](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml) and environment variables for secrets.

Sources: [Dockerfile#L1-L39](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Dockerfile#L1-L39) [README.md#L108-L115](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L108-L115)

## Kubernetes Deployment

### Deployment Configuration

The service deploys to Kubernetes using standard deployment patterns with ConfigMaps and Secrets:

### Kubernetes Resources Architecture

### Helm Chart Deployment

The service includes Helm charts for standardized Kubernetes deployments:

```shellscript
# Install using Helm
helm install websocket-service ./charts/websocket-service \
  --set image.tag=latest \
  --set environment=production \
  --set secrets.authSecret=$WEBSOCKET_AUTH_SECRET

# Upgrade deployment
helm upgrade websocket-service ./charts/websocket-service \
  --set image.tag=v1.2.3
````

### Resource Requirements

Production deployments should configure appropriate resource limits and requests:

| Resource | Request | Limit | Purpose |
| :--- | :--- | :--- | :--- |
| CPU | 100m | 500m | WebSocket handling |
| Memory | 128Mi | 512Mi | Connection state |
| Ephemeral Storage | 1Gi | 2Gi | Logs and temp files |

Sources: [README.md\#L116-L116](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L116-L116) [config/production.yaml\#L4-L8](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L4-L8)

## Configuration Management

### Environment-Specific Configuration

The service uses YAML configuration files for environment-specific settings:

### Configuration Hierarchy

### Key Configuration Parameters

Production deployments must configure the following critical parameters from [`config/production.yaml`](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml):

| Parameter | Value | Purpose |
| :--- | :--- | :--- |
| `websocket.auth.required` | `true` | Enable authentication |
| `security.cors_enabled` | `true` | Enable CORS protection |
| `security.rate_limit_enabled` | `true` | Enable rate limiting |
| `delta.reconnect_max` | `20` | Connection resilience |
| `metrics.enabled` | `true` | Observability |

Sources: [config/production.yaml\#L10-L53](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L10-L53) [README.md\#L73-L80](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L73-L80)

## Security Configuration

### Authentication and Authorization

Production deployments require proper security configuration:

```yaml
# Authentication configuration
websocket:
  auth:
    required: true
    secret: "${WEBSOCKET_AUTH_SECRET}"

# CORS configuration  
security:
  cors_enabled: true
  cors_allowed_origins: "[https://cryptovate.com](https://cryptovate.com),[https://api.cryptovate.com](https://api.cryptovate.com)"

# Rate limiting
  rate_limit_enabled: true
  rate_limit_requests: 500
  rate_limit_duration: 60
```

### TLS Configuration

The service relies on upstream TLS termination at the API Gateway or load balancer level. Ensure proper certificate management for:

  * WebSocket connections (`wss://` protocol)
  * gRPC communications (TLS encryption)
  * Health check endpoints (HTTPS)

### Secret Management

Critical secrets must be managed securely:

| Secret | Purpose | Source |
| :--- | :--- | :--- |
| `WEBSOCKET_AUTH_SECRET` | WebSocket authentication | Kubernetes Secret |
| TLS certificates | HTTPS/WSS encryption | Certificate manager |
| API keys | External service auth | Secret store |

Sources: [config/production.yaml\#L16-L26](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L16-L26) [README.md\#L124-L124](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L124-L124)

## Monitoring and Observability

### Metrics Collection

The service exposes Prometheus-compatible metrics at the `/metrics` endpoint:

### Monitoring Architecture

### Health Checks

The service provides health check endpoints for Kubernetes probes:

  * **Liveness probe**: `/health` - Basic service health
  * **Readiness probe**: `/ready` - Service readiness to handle traffic
  * **Startup probe**: `/startup` - Initial service startup validation

### Logging Configuration

Production logging uses structured JSON format with configurable levels:

```yaml
log_level: info  # production setting
```

Sources: [config/production.yaml\#L49-L53](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L49-L53) [config/production.yaml\#L6-L6](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L6-L6) [README.md\#L126-L126](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L126-L126)

## Production Best Practices

### High Availability

Deploy multiple instances with proper load balancing:

```yaml
# Kubernetes deployment
replicas: 3
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1
    maxUnavailable: 0
```

### Performance Optimization

Configure WebSocket parameters for production load:

| Parameter | Production Value | Purpose |
| :--- | :--- | :--- |
| `read_buffer_size` | 8192 | Optimize read performance |
| `write_buffer_size` | 8192 | Optimize write performance |
| `max_message_size` | 16384 | Prevent memory exhaustion |
| `rate_limit_requests` | 500 | Prevent abuse |

### Connection Management

Monitor and configure connection limits:

  * Set appropriate connection timeouts
  * Implement graceful shutdown procedures
  * Monitor connection pool utilization
  * Configure automatic reconnection parameters

### Error Handling

Implement comprehensive error handling:

  * External service connectivity failures
  * WebSocket connection drops
  * Rate limit exceeded scenarios
  * Authentication failures

Sources: [config/production.yaml\#L10-L26](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L10-L26) [config/production.yaml\#L47-L47](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L47-L47) [README.md\#L122-L128](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L122-L128)

## Troubleshooting

### Common Deployment Issues

| Issue | Symptoms | Resolution |
| :--- | :--- | :--- |
| External connectivity | Delta Exchange connection failures | Check network policies, firewall rules |
| Authentication | WebSocket connection rejections | Verify `WEBSOCKET_AUTH_SECRET` configuration |
| CORS errors | Browser WebSocket failures | Update `cors_allowed_origins` setting |
| Rate limiting | Client connection throttling | Adjust `rate_limit_requests` parameter |

### Diagnostic Commands

```shellscript
# Check pod status
kubectl get pods -l app=websocket-service

# View logs
kubectl logs -f deployment/websocket-service

# Test WebSocket connectivity
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" \
  http://service-endpoint/ws

# Check metrics
curl http://service-endpoint/metrics
```

### Performance Monitoring

Monitor key performance indicators:

  * WebSocket connection count and duration
  * Message throughput and latency
  * Error rates and types
  * Resource utilization (CPU, memory)
  * External service connectivity status

Sources: [config/production.yaml\#L29-L47](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/config/production.yaml#L29-L47) [README.md\#L126-L127](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/README.md#L126-L127)