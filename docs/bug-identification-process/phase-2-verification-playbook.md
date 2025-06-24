# Phase 2: Deep-Dive Bug-Hunting Playbook  

**Version:** 1.0  
**Target:** WebSocket Integration Service  
**Created:** $(date)  

> *Phase 2 digs below the surface uncovered in Phase 1, focusing on protocol correctness, data integrity, advanced concurrency, chaos resilience, security fuzzing, and performance scalability.*

---

## Overview & Execution Rules

1. **Sequential Execution** – Complete each Phase 2.x before moving to the next.  
2. **Documentation Requirement** – Each sub-phase must generate a summary report.  
3. **Bug Discovery Protocol** – Any discovered bug ➜ create a markdown file in `phase-2-bugs/` using the global index sequence (starting at **09**).  
4. **Index Management** – Continue numbering from Phase 1 (`08-bug-08-…`).  

---

## Phase 2.1: Protocol-Compliance Verification

**Objective:** Ensure the service exactly follows Delta Exchange's WebSocket specification and internal proto contracts.

### Steps
```bash
# 1. Extract the official spec (HTML → JSON)
python3 tools/extract_delta_spec.py docs/delta-exchange-api-docs/whole-html.html delta_spec.json

# 2. Generate contract tests from the spec
python3 tools/generate_contract_tests.py delta_spec.json > contract_tests.go

go test ./tests/contract -v | tee protocol-compliance.log
```

### Success Criteria
- All generated contract tests pass.
- No schema mismatches, missing fields, or wrong types.

### Failure Indicators
- Test failures in `protocol-compliance.log`.
- Divergent message formats, missing subscription acknowledgements.

### Bug Categories to Check
- Message-format deviations  
- Unsupported/unhandled server opcodes  
- Incorrect error-response semantics

---

## Phase 2.2: Data-Integrity Tracing

**Objective:** Detect corruption, precision loss, or transformation errors along the data pipeline (Delta → Client → Handler → WebSocket).

### Steps
```bash
# Instrument transformation checkpoints (trace mode)
WEBSOCKET_TRACE=1 make run &
SERVICE_PID=$!

# Pump reference data through the system
python3 tools/replay_recorded_delta_frames.py tests/fixtures/ticker.json

sleep 10; kill $SERVICE_PID
```

### Success Criteria
- Checksums identical before & after each transformation.
- JSON schema untouched, numerical precision preserved.

### Failure Indicators
- Checksum mismatch lines in `trace-*.log`.
- Float precision drift > 1e-8.

### Bug Categories
- Truncation/rounding errors  
- UTF-8 encoding issues  
- Dropped/duplicated fields

---

## Phase 2.3: Advanced Concurrency & Race Analysis

**Objective:** Uncover complex race conditions, deadlocks, message ordering issues.

### Steps
```bash
# Build with race detector and stress test
GOMAXPROCS=4 go test -race -run TestHighConcurrency -count=500 ./... | tee race-analysis.log

go run -race cmd/loadtest/main.go -connections 500 -duration 2m | tee load-race.log
```

### Success Criteria
- No `DATA RACE`, deadlock, or panic in logs.

### Failure Indicators
- `fatal error: concurrent map read and map write`.
- Goroutine leaks > baseline + 5 % after test.

### Bug Categories
- Deadlocks during shutdown  
- Lost messages under contention  
- Starvation of write pumps

---

## Phase 2.4: Resource-Exhaustion & Memory Profiling

**Objective:** Detect leaks and abnormal resource growth under sustained load.

### Steps
```bash
make run &
SERVICE_PID=$!
# 5 min load simulation
python3 tools/ws_loadgen.py --conn 1000 --msgs 5000 --duration 300

go tool pprof -top http://localhost:8080/debug/pprof/heap > heap_profile.txt
kill $SERVICE_PID
```

### Success Criteria
- Stable RSS & goroutine count (±10 %).
- No continuous growth trend in heap profile.

### Failure Indicators
- Unbounded goroutine or memory growth.
- File-descriptor exhaustion (> 90 % soft limit).

### Bug Categories
- Channel/connection leaks  
- Unbounded buffering  
- Improper defer/Close logic

---

## Phase 2.5: Chaos / Fault-Injection Resilience

**Objective:** Validate graceful degradation under partial failures.

### Steps
```bash
# Introduce 100 ms latency & 2 % packet loss to Delta traffic
sudo tc qdisc add dev lo root netem delay 100ms loss 2%

make run &; SERVICE_PID=$!

# Chaos script: random kill -STOP/-CONT, network drops
python3 tools/chaos_monkey.py --pid $SERVICE_PID --duration 120

sudo tc qdisc del dev lo root netem
```

### Success Criteria
- Service recovers without manual intervention.
- No unhandled panics; reconnect logic respects `reconnect_max`.

### Failure Indicators
- Service stuck in reconnect loop.
- Panic or memory explosion during chaos.

### Bug Categories
- Improper back-off logic  
- Un-guarded nil pointer deref after reconnect  
- State corruption on partial failures

---

## Phase 2.6: Security Fuzzing & Attack Simulation

**Objective:** Uncover input-validation bugs, auth bypasses, DoS vectors.

### Steps
```bash
# WebSocket JSON fuzzing
cd fuzz && go test -fuzz=FuzzWSMessage -fuzztime=2m | tee fuzz.log

# Auth bypass attempts
python3 tools/auth_bypass_probe.py ws://localhost:8080/ws
```

### Success Criteria
- Fuzzing produces **no** crashes or data races.
- All unauthenticated requests rejected when auth required.

### Failure Indicators
- Go panic during fuzz run.
- Auth bypass script obtains 200 responses.

### Bug Categories
- JSON unmarshal panics  
- Infinite-loop DoS  
- Missing rate-limit enforcement

---

## Phase 2.7: Performance & Scalability Benchmarking

**Objective:** Ensure throughput and latency meet SLOs under realistic peak load.

### Steps
```bash
make run &; SERVICE_PID=$!

# Throughput benchmark (100 k msgs/s target)
artillery run tests/perf/ws_throughput.yml > perf_report.json

kill $SERVICE_PID
```

### Success Criteria
- P95 WebSocket latency < 50 ms.
- Max CPU utilisation < 75 % of allotted cores.

### Failure Indicators
- Latencies or CPU breaching SLO.
- Garbage-collection pauses > 100 ms.

### Bug Categories
- Lock contention hot-spots  
- Inefficient JSON marshaling  
- Unnecessary reflection in hot path

---

## Phase Completion Protocol

After each sub-phase:
1. **Generate Phase Summary** → `phase-2-summary-2.x.md` under `phase-2-bugs/`
2. **Update Bug Index** → `phase-2-bugs/00-overview_of_bugs.md`
3. **Clean Artifacts** – remove heavy logs, pprof dumps
4. **Commit** – docs & bug files with descriptive messages

---

## Success Metrics

| Metric | Target | Critical Threshold |
| ------ | ------ | ------------------ |
| New critical bugs | Document all | ≥1 triggers immediate fix |
| Unresolved crashes | 0 | Any crash = fail |
| Memory leak | 0 bytes/min growth | >256 KB/min for ≥3 min |
| P95 latency | <50 ms | >100 ms |
| Max CPU | <75 % cores | ≥90 % for 30 s |

---

**End of Phase 2 Playbook** 