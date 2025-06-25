# Websocket-Service Telemetry Integration Plan

> This document converts the high-level observability blueprint into a **step-by-step, file-oriented work queue** that any contributor can follow to wire OpenTelemetry tracing, metrics and structured logging into the code-base.
>
>  • Everything is grouped by **phase** – you can finish and commit each phase independently.  
>  • Every instrumentation item lists **file / function / insertion line(s) / justification** so that reviews are fast and no "where exactly?" discussions are needed.  
>  • The chosen stack (OpenTelemetry SDK → Prometheus exporter → Jaeger/Tempo) keeps us 100 % **vendor-neutral** while matching the tooling already available in K8s clusters.

---
## 0. Chosen Stack & Why

| Layer | Choice | Rationale |
|-------|--------|-----------|
| **Instrumentation API** | OpenTelemetry-Go v1.21 | CNCF standard → out-of-the-box exporters, W3C trace-context, community momentum; no licence fees. |
| **Metrics Export** | `go.opentelemetry.io/contrib/exporters/metric/prometheus` | Emits native Prometheus format that existing platform Prometheus scrapes; keeps one metrics backend. |
| **Tracing Export** | Jaeger **or** Grafana Tempo via OTLP | Both accept OTLP/HTTP without agents; switch is a Helm value, no code change. |
| **Logging** | OTel `semconv` structured logs + `zap` JSON encoder | Single schema, trace-id injection, easy FluentBit shipping. |
| **Propagation** | W3C Trace-Context (`traceparent`) | Supported by browsers, Envoy, gRPC, JS clients → end-to-end correlation. |

> **No vendor lock-in**: the only compile-time deps are opentelemetry-go & prometheus client; switching exporters = config change.

### `go.mod` lines (add once)
```
require (
	go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/sdk v1.34.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/exporters/prometheus v0.58.0
	github.com/prometheus/client_golang v1.20.5
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0
)
```

---
## 1. Phased Implementation Road-map

| Phase | Goal | Outcome |
|-------|------|---------|
| **P-1 Core Bootstrap** | Provide one `telemetry` package that initialises global tracer / meter, Prometheus handler, Jaeger exporter. | Binary starts with `OTEL_…` env vars and exposes `/metrics` unchanged. |
| **P-2 Delta Client Ingest** | Instrument Delta frame reception & JSON parsing. | Metrics `delta_messages_total`, span `delta.receive`. |
| **P-3 Broadcast Hub** | Instrument `WebsocketHandler.BroadcastToChannel`. | `broadcast_total`, `broadcast_latency_ms`, span `broadcast.filter`. |
| **P-4 Client Delivery** | Instrument `writePump` per message. | `client_delivery_total`, `client_bytes_sent`. |
| **P-5 gRPC Server** | Wrap server with `otelgrpc`. | Automatic RPC spans + `grpc_server_duration_seconds`. |
| **P-6 Error/Panic** | Capture `recover()` panics + log → `go_panic_total`. | Crash root cause surfaced. |
| **P-7 Dashboards & Alerts** | Grafana dashboards, Prom rules. | Ready-made panels & SLO alerts. |

Each phase is independent; merge on green CI.

---
## 1.1 Implementation Protocol & Quality Gates (read **before** starting any phase)

1. **One-phase-per-PR** – never mix items from different phases in the same branch.
2. **Branch naming**: `telemetry/phase-<n>` (e.g. `telemetry/phase-1-core-bootstrap`).
3. **Trigger phrase** in PR or chat to CODEX: `IMPLEMENT THE PHASE-<n> OF LOGGING in docs/telemetry-planning.md` – spins up the automated workflow for that phase.
4. **Double-check checklist** (CI enforces):
   - `make ci` green (build, vet, lint, tests, race).
   - Added / updated unit & integration tests prove signals are emitted.
   - `/metrics` endpoint and Jaeger traces verified in a smoke test.
   - Documentation sections & checkboxes in this file updated.
5. **Rollback ready** – guard each new exporter / handler behind a feature flag (`TELEMETRY_PHASE_<n>_ENABLED`) so it can be disabled at runtime.
6. **Reviewer focus** – reviewers must check for cardinality explosions, context propagation leaks and exporter error handling.

---
## 2. Instrumentation Work-Items (file oriented)

> **Line numbers** refer to current *dev-phase-02* branch snapshot. Caution: re-check after rebases.

| 📄 **File** | 🔧 **Function** | 📍 **Insert @ line** | 🚀 **What to add** | 💡 **Why (insight gained)** |
|-------------|----------------|--------------------|-------------------|----------------------------|
| `main.go` | `main()` | ~30 (after config load) | Call `telemetry.Init(ctx, cfg)` | Boot global tracer / meter once. |
| `telemetry/init.go` *(new)* | `Init` | – | build tracerProvider, Prom exporter, Jaeger exporter, `/metrics` mux, set global logger. | Foundation for all other spans/metrics. |
| `internal/clients/delta_websocket.go` | `readPump()` | 140-175 | `ctx, span := tracer.Start(ctx, "delta.receive")` before `conn.ReadMessage()`; record message size attr; `meter.Counter("delta_messages_total")`. | Answers: *Is Delta alive? How big?* Helps bug #07 (metrics missing). |
| same file | `json.Unmarshal` error path | 150-155 | `jsonUnmarshalErr.Add(ctx, 1)` | Surfaces malformed JSON (#11). |
| `internal/handlers/websocket_handler.go` | `BroadcastToChannel` | 135-180 | Start span `broadcast.filter`; histogram timer around loop; increment `broadcast_total` counter. | Detect hot channels and latency spikes (#03/#09). |
| same file | `writePump()` | 365-405 | span `client.write` per flush; counters `client_delivery_total{result}` and `client_bytes_sent`. | See slow / blocked clients and back-pressure. |
| same file | `run()` default case | 290-310 | when select default triggers client drop → `broadcast_drop_total` + warn log with trace id. | Quantify dropped messages (#03/#10). |
| `internal/server/server.go` | `NewServer()` | before returning srv | wrap gRPC server: `grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))` | Auto traces + metrics for public API. |
| `cmd/otel-demo` *(new)* | – | – | simple docker-compose file with Jaeger & Prometheus for local test. | Developer UX. |

*(Use `telemetry/meter.go` helpers for counter/histogram creation to avoid duplication.)*

---
## 3. Correlation-ID & Logging Strategy

1. **Generate ULID** for every Delta frame (cheap & sortable).  
2. Store in `context.Context` with helper `telemetry.WithMsgID(ctx, id)`; retrieve via `telemetry.MsgID(ctx)`.  
3. Inject into span attributes **and** JSON pushed to client (`"trace_id"`, `"x_msg_id"`).  
4. Structured logs emitted through `zap` contain `trace_id` → one-click pivot between logs ↔ traces in Grafana Loki.

---
## 4. Sampling & Cardinality Guard-rails

* Head-based 5 % trace sampling enabled in `Init()`; tunable by `OTEL_TRACES_SAMPLER_ARG`.  
* **Metrics labels capped**: `product_id` more than 20 distinct → record as `_other`.  
* Client IDs **never** used as label, only as span attr.

---
## 5. Testing Hooks

| Test | Method | Expected signal |
|------|--------|-----------------|
| Unit | Inject fake frame → `readPump` | `delta_messages_total` == 1 |
| Integration | Run service with `OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318` | Spans visible in Jaeger UI |
| Fuzz | Replay invalid JSON | `json_unmarshal_errors_total` > 0 |
| Load | 1 k broadcasts/s | `broadcast_latency_ms` P95 < 50 ms |

---
## 6. Work-Item Checklist

- [x] **Create telemetry package** (`telemetry/init.go`, `ctx.go`, `meter.go`).
- [x] Add dependencies to `go.mod` & `go.sum`.
- [x] **Phase 1** – wire `main.go` bootstrap, expose `/metrics`.
 - [x] **Phase 2** – instrument `DeltaWebsocketClient.readPump` (span + counters).
- [x] **Phase 3** – instrument `BroadcastToChannel` (span, counter, hist).
- [x] **Phase 4** – instrument `writePump` (span, counters, bytes).
- [x] **Phase 4b** – error counter in `run()` default drop path.
- [ ] **Phase 5** – wrap gRPC server with `otelgrpc` interceptor.
- [ ] **Phase 6** – panic recovery middleware + `go_panic_total`.
- [ ] Build & run **docker-compose telemetry demo**; validate metrics & traces.
- [ ] Commit dashboards JSON & PrometheusRule manifests to `deploy/telemetry/`.

> **When every box is ticked, telemetry work is complete and we can close bugs #05 (port mismatch clues via metrics), #07 (metrics endpoint), #11 (JSON errors surfaced) and guard against #03/#10 concurrency panics.** 