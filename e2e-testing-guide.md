# End-to-End Testing Guide for WebSocket Service

## Pre-Testing Setup

### 1. Build and Start the Service
```bash
# Clean build to ensure everything is fresh
make clean
make build

# Start the service
make run
```

### 2. Verify Service is Running
```bash
# Check if both servers started successfully
curl -s http://localhost:8080/health
# Expected: "OK"

# Check if metrics endpoint is accessible
curl -s http://localhost:8080/metrics | head -5
# Expected: Prometheus metrics format

# Check gRPC server health (requires grpcurl)
grpcurl -plaintext localhost:9090 list
# Expected: List of gRPC services
```

---

## 🧪 Test Categories

### A. HTTP Server Functionality

#### A1. Health Check Endpoint
```bash
# Basic health check
curl -i http://localhost:8080/health

# Expected Response:
# HTTP/1.1 200 OK
# Content-Length: 2
# 
# OK
```

#### A2. Metrics Endpoint
```bash
# Check metrics are being collected
curl -s http://localhost:8080/metrics | grep -E "websocket|grpc|http"

# Expected: Various Prometheus metrics related to the service
```

#### A3. CORS and Security Headers
```bash
# Test CORS preflight request
curl -i -X OPTIONS http://localhost:8080/ws \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Connection,Upgrade"

# Expected: Appropriate CORS headers if enabled in config
```

### B. WebSocket Functionality

#### B1. WebSocket Connection Test
```bash
# Install wscat if not available
npm install -g wscat

# Test basic WebSocket connection
wscat -c ws://localhost:8080/ws

# You should see: Connected (press CTRL+C to quit)
```

#### B2. Subscription Management
Once connected via wscat, test these message flows:

**Subscribe to a channel:**
```json
{
  "type": "subscribe",
  "payload": {
    "channels": [
      {
        "name": "v2/ticker",
        "symbols": ["BTC_USDT", "ETH_USDT"]
      }
    ]
  }
}
```

**Expected Response:**
```json
{
  "type": "subscription_confirmation",
  "channel": "v2/ticker",
  "status": "subscribed"
}
```

**Unsubscribe from a channel:**
```json
{
  "type": "unsubscribe",
  "payload": {
    "channels": [
      {
        "name": "v2/ticker"
      }
    ]
  }
}
```

**Ping/Pong test:**
```json
{
  "type": "ping"
}
```

**Expected Response:**
```json
{
  "type": "pong"
}
```

#### B3. Multi-Client Subscription Test
```bash
# Terminal 1
wscat -c ws://localhost:8080/ws

# Terminal 2
wscat -c ws://localhost:8080/ws

# Terminal 3
wscat -c ws://localhost:8080/ws

# In each terminal, subscribe to the same channel and verify all receive data
```

### C. gRPC API Testing

Install grpcurl if not available:
```bash
# macOS
brew install grpcurl

# Ubuntu/Debian
sudo apt-get install grpcurl
```

#### C1. List Available Services
```bash
grpcurl -plaintext localhost:9090 list
# Expected: websocket.v1.WebsocketService
```

#### C2. Get Service Statistics
```bash
grpcurl -plaintext localhost:9090 \
  websocket.v1.WebsocketService/GetStatistics

# Expected JSON response with:
# - active_connections
# - active_subscriptions  
# - messages_sent
# - messages_received
# - external_sources status
```

#### C3. Get Connection Status
```bash
grpcurl -plaintext localhost:9090 \
  websocket.v1.WebsocketService/GetConnectionStatus

# Expected: Status of Delta Exchange connection
```

#### C4. Broadcast Message
```bash
# First, have a WebSocket client connected and subscribed
# Then broadcast a message via gRPC
grpcurl -plaintext -d '{
  "channel": "v2/ticker",
  "message": "dGVzdCBtZXNzYWdl",
  "product_ids": ["BTC_USDT"]
}' localhost:9090 websocket.v1.WebsocketService/Broadcast

# The connected WebSocket client should receive the message
```

### D. Configuration Testing

#### D1. Environment Variable Overrides
```bash
# Stop the current service (Ctrl+C)

# Test port overrides
HTTP_PORT=8888 GRPC_PORT=9999 make run

# In another terminal, verify new ports
curl -s http://localhost:8888/health
grpcurl -plaintext localhost:9999 list
```

#### D2. Configuration File Loading
```bash
# Test with different config files
cp config/local.yaml config/test.yaml
# Modify some values in test.yaml

# Start with test config
ENVIRONMENT=test make run
```

### E. Delta Exchange Integration

#### E1. External Connection Verification
```bash
# Check if service connects to Delta Exchange
curl -s http://localhost:8080/metrics | grep delta

# Check connection status via gRPC
grpcurl -plaintext localhost:9090 \
  websocket.v1.WebsocketService/GetConnectionStatus
```

#### E2. Data Flow Verification
```bash
# Connect WebSocket client and subscribe to ticker
wscat -c ws://localhost:8080/ws

# Send subscription for configured symbols
{
  "type": "subscribe",
  "payload": {
    "channels": [
      {
        "name": "v2/ticker", 
        "symbols": ["BTC_USDT"]
      }
    ]
  }
}

# Wait to receive real-time ticker data from Delta Exchange
# Expected: JSON messages with ticker updates
```

### F. Telemetry and Monitoring

#### F1. OpenTelemetry Metrics
```bash
# Check Prometheus metrics are being generated
curl -s http://localhost:8080/metrics | grep -E "otel|telemetry"

# Look for WebSocket-specific metrics
curl -s http://localhost:8080/metrics | grep -E "websocket_|grpc_|http_"
```

#### F2. Error Handling and Recovery
```bash
# Test panic recovery middleware
curl -X POST http://localhost:8080/nonexistent-endpoint

# Check logs for proper error handling
```

### G. Load and Stress Testing

#### G1. Concurrent WebSocket Connections
```bash
# Create multiple concurrent connections
for i in {1..20}; do
  wscat -c ws://localhost:8080/ws &
done

# Check connection stats
grpcurl -plaintext localhost:9090 \
  websocket.v1.WebsocketService/GetStatistics
```

#### G2. High-Frequency Message Testing
```bash
# Connect client and send rapid subscription/unsubscription
wscat -c ws://localhost:8080/ws

# Script to send multiple rapid messages
for i in {1..100}; do
  echo '{"type":"ping"}' | wscat -c ws://localhost:8080/ws
done
```

### H. Race Condition and Concurrency Testing

#### H1. Race Detector Testing
```bash
# Build with race detector
go build -race -o websocket-service-race main.go

# Run with race detection
./websocket-service-race &

# Run concurrent operations
go test -race ./tests/bugs/... -v

# Expected: No race condition warnings
```

#### H2. Concurrent Subscription Operations
```bash
# Run the reproduction tests that verify race condition fixes
go test -race ./tests/bugs -run TestBug03_Repro -v
go test -race ./tests/bugs -run TestBug09_Repro -v  
go test -race ./tests/bugs -run TestBug10_Repro -v

# All should pass without race warnings
```

---

## 🎯 End-to-End Integration Test Scenarios

### Scenario 1: Complete User Journey
1. **Client connects** to WebSocket endpoint
2. **Subscribes** to ticker channel for BTC_USDT
3. **Receives real-time data** from Delta Exchange
4. **Service broadcasts** data to client
5. **Client unsubscribes** from channel
6. **Verifies** Delta connection is managed properly

### Scenario 2: Multi-Client Resource Management
1. **Connect 10 clients** simultaneously
2. **All subscribe** to same channel
3. **Verify** all receive same data
4. **Disconnect 9 clients**
5. **Verify** Delta subscription maintained for remaining client
6. **Disconnect last client**
7. **Verify** Delta unsubscribes to save resources

### Scenario 3: Service Resilience
1. **Start service** with invalid Delta configuration
2. **Verify** service starts but reports connection errors
3. **Fix configuration** via environment variables
4. **Restart service**
5. **Verify** Delta connection establishes successfully

---

## 🔍 Verification Checklist

### Core Functionality ✅
- [ ] HTTP server starts on configured port
- [ ] gRPC server starts on configured port  
- [ ] Health endpoint responds correctly
- [ ] Metrics endpoint provides Prometheus format
- [ ] WebSocket connections can be established
- [ ] Subscribe/unsubscribe messages work correctly
- [ ] Ping/pong mechanism functions
- [ ] Multiple clients can connect simultaneously

### Integration ✅
- [ ] Delta Exchange connection establishes
- [ ] Real-time data flows from Delta to clients
- [ ] Subscription management affects Delta connection
- [ ] External connection status is accurate

### Configuration ✅
- [ ] YAML configuration loads correctly
- [ ] Environment variable overrides work
- [ ] Port configuration is respected
- [ ] Delta configuration is applied

### Telemetry ✅
- [ ] OpenTelemetry metrics are generated
- [ ] Prometheus metrics endpoint works
- [ ] Error recovery middleware functions
- [ ] gRPC telemetry is collected

### Performance ✅
- [ ] Service handles concurrent connections
- [ ] No race conditions under load
- [ ] Memory usage stays reasonable
- [ ] Response times are acceptable

### Error Handling ✅
- [ ] Invalid messages are handled gracefully
- [ ] Network failures don't crash service
- [ ] Panic recovery works correctly
- [ ] Proper error messages in logs

---

## 🚨 Quick Automated Test

Run this comprehensive test:

```bash
# Full CI pipeline
make ci

# All bug regression tests
go test -race ./tests/bugs/... -v

# Integration tests
go test -race ./... -v

# Expected: All tests pass, no race conditions
```

---

## 📋 Manual Testing Workflow

1. **Start the service**: `make run`
2. **Open 3 terminals** for parallel testing
3. **Terminal 1**: WebSocket client (`wscat -c ws://localhost:8080/ws`)
4. **Terminal 2**: gRPC testing (`grpcurl` commands)
5. **Terminal 3**: HTTP endpoint testing (`curl` commands)
6. **Follow test scenarios** A through H above
7. **Verify all expected behaviors**
8. **Check logs** for any errors or warnings

This comprehensive testing plan ensures every component of your WebSocket service is functioning correctly after the bug fixes! 