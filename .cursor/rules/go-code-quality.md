# Go Code Quality Rules

These rules ensure consistent, high-quality Go code throughout the WebSocket service.

## Error Handling

### Always handle errors explicitly
```go
// ❌ Bad: Ignoring errors
result, _ := someFunction()

// ✅ Good: Explicit error handling
result, err := someFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Wrap errors with context
```go
// ❌ Bad: Generic error return
if err != nil {
    return err
}

// ✅ Good: Contextual error wrapping
if err != nil {
    return fmt.Errorf("failed to establish websocket connection: %w", err)
}
```

### Use error variables for expected errors
```go
// ✅ Good: Define error variables
var (
    ErrConnectionClosed = errors.New("websocket connection closed")
    ErrInvalidMessage   = errors.New("invalid message format")
)
```

## Resource Management

### Always use defer for cleanup
```go
// ✅ Good: Ensure resources are cleaned up
func handleConnection(conn *websocket.Conn) error {
    defer conn.Close()
    
    // Connection handling logic
    return nil
}
```

### Check for nil before dereferencing
```go
// ❌ Bad: Potential nil pointer dereference
func processClient(client *Client) {
    client.Send(message) // Could panic if client is nil
}

// ✅ Good: Nil check
func processClient(client *Client) error {
    if client == nil {
        return ErrNilClient
    }
    return client.Send(message)
}
```

## Concurrency

### Protect shared state with mutexes
```go
// ✅ Good: Thread-safe map operations
type ClientManager struct {
    mu      sync.RWMutex
    clients map[string]*Client
}

func (cm *ClientManager) AddClient(id string, client *Client) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    cm.clients[id] = client
}
```

### Use contexts for cancellation
```go
// ✅ Good: Context-aware operations
func (h *Handler) processMessages(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg := <-h.messageChan:
            if err := h.handleMessage(msg); err != nil {
                return err
            }
        }
    }
}
```

### Properly manage goroutines
```go
// ✅ Good: Goroutine lifecycle management
func (s *Server) Start(ctx context.Context) error {
    var wg sync.WaitGroup
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        s.handleConnections(ctx)
    }()
    
    <-ctx.Done()
    wg.Wait()
    return nil
}
```

## Configuration

### Use struct tags for configuration
```go
// ✅ Good: Clear configuration structure
type Config struct {
    HTTPPort     int           `yaml:"http_port" env:"HTTP_PORT" default:"8080"`
    GRPCPort     int           `yaml:"grpc_port" env:"GRPC_PORT" default:"9090"`
    ReadTimeout  time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" default:"30s"`
    WriteTimeout time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" default:"30s"`
}
```

### Validate configuration on startup
```go
// ✅ Good: Configuration validation
func (c *Config) Validate() error {
    if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
        return fmt.Errorf("invalid HTTP port: %d", c.HTTPPort)
    }
    if c.ReadTimeout <= 0 {
        return fmt.Errorf("read timeout must be positive: %v", c.ReadTimeout)
    }
    return nil
}
```

## Logging

### Use structured logging
```go
// ❌ Bad: Unstructured logging
log.Printf("Client connected from %s", addr)

// ✅ Good: Structured logging with context
logger.Info("client connected",
    "remote_addr", addr,
    "connection_id", connID,
    "timestamp", time.Now(),
)
```

### Log at appropriate levels
```go
// ✅ Good: Appropriate log levels
logger.Debug("processing subscription", "channel", channel)  // Debug info
logger.Info("client connected", "id", clientID)             // Normal operations
logger.Warn("retry attempt", "attempt", retryCount)         // Warnings
logger.Error("connection failed", "error", err)             // Errors
```

## Code Organization

### Use meaningful package names
```go
// ❌ Bad: Generic package names
package utils
package helpers

// ✅ Good: Descriptive package names
package handlers
package websocket
```

### Group related functionality
```go
// ✅ Good: Logical grouping
type WebsocketHandler struct {
    // Connection management
    connections map[string]*Connection
    mu          sync.RWMutex
    
    // Message handling
    broadcaster *MessageBroadcaster
    router      *MessageRouter
    
    // Configuration
    config *Config
    logger *Logger
}
```

## Testing

### Write table-driven tests
```go
// ✅ Good: Table-driven test structure
func TestMessageValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid subscribe", `{"type":"subscribe","channel":"ticker"}`, false},
        {"invalid json", `{"invalid":}`, true},
        {"missing type", `{"channel":"ticker"}`, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateMessage(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateMessage() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Use testify for assertions
```go
// ✅ Good: Clear test assertions
func TestConnectionHandling(t *testing.T) {
    handler := NewWebsocketHandler(config)
    
    conn := &MockConnection{}
    err := handler.HandleConnection(conn)
    
    assert.NoError(t, err)
    assert.True(t, conn.WasClosed())
}
```

## Performance

### Use sync.Pool for frequent allocations
```go
// ✅ Good: Object pooling for performance
var messagePool = sync.Pool{
    New: func() interface{} {
        return &Message{}
    },
}

func getMessage() *Message {
    return messagePool.Get().(*Message)
}

func putMessage(msg *Message) {
    msg.Reset()
    messagePool.Put(msg)
}
```

### Prefer channels over locks when possible
```go
// ✅ Good: Channel-based coordination
type MessageProcessor struct {
    inbound  chan Message
    outbound chan Message
    quit     chan struct{}
}

func (mp *MessageProcessor) Process() {
    for {
        select {
        case msg := <-mp.inbound:
            processed := mp.transform(msg)
            mp.outbound <- processed
        case <-mp.quit:
            return
        }
    }
}
```

## Security

### Validate input data
```go
// ✅ Good: Input validation
func validateSubscriptionRequest(req *SubscriptionRequest) error {
    if req.Channel == "" {
        return ErrEmptyChannel
    }
    if len(req.ProductIDs) > MaxProductIDs {
        return ErrTooManyProductIDs
    }
    for _, id := range req.ProductIDs {
        if id <= 0 {
            return fmt.Errorf("invalid product ID: %d", id)
        }
    }
    return nil
}
```

### Use constants for sensitive values
```go
// ✅ Good: Constants for limits and sensitive values
const (
    MaxConnections     = 1000
    MaxMessageSize     = 1024 * 1024 // 1MB
    ConnectionTimeout  = 30 * time.Second
    HeartbeatInterval  = 15 * time.Second
)
```

These rules should be applied consistently throughout the codebase to maintain quality and prevent common Go pitfalls. 