# Resolving Bugs

## Bugs

### 1. Wrong default ports in hard-coded config
The fallback config in `internal/config/config.go` sets `HTTPPort` **8083** and `GRPCPort` **9093** instead of the documented 8080 / 9090. Running the binary without external YAML therefore starts the servers on unexpected ports.

**Fix idea**: Change the defaults in `LoadConfig` to 8080 and 9090 or load the YAML configs even in local mode.

### 2. `make build` fails – protobuf code not generated
Running `make build` on a fresh clone fails with:
```
main.go:14:2: no required module provides package github.com/Cryptovate-India/websocket-service/gen/websocket/api/v1
```
The `build` target doesn't depend on the `proto` target, so generated code is missing.

**Quick fix**
Add `proto` as a prerequisite:
```makefile
build: proto
	$(GO) build $(LDFLAGS) -o $(SERVICE_NAME) main.go
```
Or run `make proto` before building.

### 3. Inconsistent Delta websocket URL across environments
`delta.url` differs between configs:
* local → `wss://socket.india.delta.exchange`
* dev / prod → `wss://socket.delta.exchange`

Using the India endpoint locally returns INR-quoted symbols (`BTCUSD`) while dev/prod return USD (`BTCUSD`). I have finalised to use the india endpoint everywhere.

### 4. Input format in the README.md file is not correct
README shows a flat `product_ids` subscribe message. Per Delta docs it should be nested under `payload.channels[].symbols`:
```json
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
```
