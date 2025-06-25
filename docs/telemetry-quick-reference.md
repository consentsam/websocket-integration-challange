# Telemetry Quick Reference

**WebSocket Service Observability Cheat Sheet**

---

## 🚀 Quick Start

```bash
# Start service with metrics
./websocket-service

# Check metrics endpoint
curl http://localhost:8080/metrics

# Monitor specific metrics
curl -s http://localhost:8080/metrics | grep "delta_messages_total\|broadcast_latency_ms"
```

---

## 📊 Key Metrics at a Glance

| Metric | Type | What it Tracks | Alert When |
|--------|------|----------------|------------|
| `delta_messages_total` | Counter | Messages from Delta Exchange | Rate drops to 0 |
| `broadcast_latency_ms` | Histogram | WebSocket broadcast time | P95 > 100ms |
| `client_delivery_total{result="error"}` | Counter | Failed client deliveries | Rate > 1% |
| `go_panic_total` | Counter | Service crashes | Any increase |
| `json_unmarshal_errors_total` | Counter | JSON parsing failures | Any increase |

---

## 🐛 Common Debug Scenarios

### **No Messages Flowing**
```bash
# Check Delta connection
curl -s http://localhost:8080/metrics | grep delta_messages_total
# Should be incrementing

# Check WebSocket broadcasts  
curl -s http://localhost:8080/metrics | grep broadcast_total
# Should match delta messages
```

### **High Latency**
```bash
# Check broadcast performance
curl -s http://localhost:8080/metrics | grep broadcast_latency_ms_bucket
# Look for high bucket values
```

### **Client Issues**
```bash
# Check delivery success rate
curl -s http://localhost:8080/metrics | grep 'client_delivery_total{result="ok"}'
curl -s http://localhost:8080/metrics | grep 'client_delivery_total{result="error"}'
# Compare rates
```

---

## ⚡ Grafana Queries

### **Service Health Dashboard**
```promql
# Uptime
up{job="websocket-service"}

# Message throughput (messages/sec)
rate(delta_messages_total[5m])

# Error rate (errors/sec)  
rate(json_unmarshal_errors_total[5m]) + rate(go_panic_total[5m])

# Broadcast latency P95
histogram_quantile(0.95, rate(broadcast_latency_ms_bucket[5m]))
```

### **Performance Dashboard**
```promql
# Client delivery success rate
rate(client_delivery_total{result="ok"}[5m]) / rate(client_delivery_total[5m])

# Bandwidth usage (bytes/sec)
rate(client_bytes_sent[5m])

# Message drop rate
rate(broadcast_drop_total[5m])
```

---

## 🚨 Critical Alerts

```yaml
# Copy-paste alert rules
- alert: ServiceDown
  expr: up{job="websocket-service"} == 0
  for: 1m

- alert: NoMessages
  expr: rate(delta_messages_total[5m]) == 0
  for: 2m

- alert: HighLatency  
  expr: histogram_quantile(0.95, rate(broadcast_latency_ms_bucket[5m])) > 100
  for: 2m

- alert: PanicDetected
  expr: increase(go_panic_total[5m]) > 0
  for: 0m
```

---

## 🔧 Environment Variables

```bash
# Change metrics endpoint
WEBSOCKET_METRICS_ENDPOINT="/custom-metrics"

# Disable metrics
WEBSOCKET_METRICS_ENABLED=false

# Change environment
ENVIRONMENT=production

# Override ports
WEBSOCKET_HTTP_PORT=9999
WEBSOCKET_GRPC_PORT=9998
```

---

## 📈 Troubleshooting Commands

```bash
# Service health check
curl -f http://localhost:8080/metrics > /dev/null && echo "✅ Metrics OK" || echo "❌ Metrics FAIL"

# Count total metrics
curl -s http://localhost:8080/metrics | grep -c "^#"

# Check for high cardinality
curl -s http://localhost:8080/metrics | grep -E "client_delivery_total|broadcast_total" | wc -l

# Monitor in real-time
watch -n 5 'curl -s http://localhost:8080/metrics | grep -E "delta_messages_total|broadcast_total"'
```

---

## 💡 Pro Tips

- **Use labels wisely**: Avoid high cardinality (no client IDs, use bounded values)
- **Monitor the monitors**: Set up alerts on metric staleness
- **Correlate with traces**: Use trace IDs in logs for better debugging
- **Histogram buckets**: Default buckets work for most latency use cases
- **Counter vs Gauge**: Counters for events, gauges for states

---

*Full documentation: [docs/telemetry-implementation.md](./telemetry-implementation.md)* 