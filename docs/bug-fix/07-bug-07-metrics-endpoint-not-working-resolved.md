# Bug 07 Resolved - Metrics Endpoint Not Working

Bug File: [docs/bugs/07-bug-07-metrics-endpoint-not-working.md](../bugs/07-bug-07-metrics-endpoint-not-working.md)

## Verification Steps
1. Build and start the service:
   ```bash
   go build -o websocket-service main.go
   ./websocket-service &
   ```
2. Query the metrics endpoint:
   ```bash
   curl -i http://localhost:8080/metrics | head
   ```
   The response should return HTTP `200 OK` and Prometheus metrics text.
