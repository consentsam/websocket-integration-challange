# Bug Identification & Verification Playbook

**Version:** 1.0  
**Target:** WebSocket Integration Service  
**Updated:** $(date)

This document provides a systematic, repeatable process for identifying bugs, documenting findings, and ensuring code quality across the websocket service codebase.

---

## Overview

This playbook follows a **7-phase approach** that progresses from static analysis to dynamic testing, ensuring comprehensive coverage of potential issues.

### Phase Execution Rules

1. **Sequential Execution**: Complete each phase before proceeding to the next
2. **Documentation Requirement**: Every phase must produce a summary report
3. **Bug Discovery Protocol**: Any discovered bug triggers immediate documentation using the standardized template
4. **Index Management**: Bugs are numbered sequentially regardless of discovery phase

---

## Phase 1: Environment & Dependency Verification

### Phase 1.1: Build System Verification
**Objective**: Ensure the codebase can be built without errors

**Steps**:
```bash
# Clean any previous builds
make clean || rm -f main

# Attempt build and capture all output
make build 2>&1 | tee build-verification.log

# Check for binary creation
ls -la main
file main
```

**Success Criteria**: 
- Build completes without errors
- Binary is created and executable
- No warning messages about missing dependencies

**Failure Indicators**:
- Compilation errors
- Missing dependencies
- Import path issues
- Version conflicts

### Phase 1.2: Dependency Analysis
**Objective**: Verify all dependencies are properly resolved

**Steps**:
```bash
# Check go.mod integrity
go mod verify

# Download and verify dependencies  
go mod download
go mod tidy

# Check for security vulnerabilities
go list -json -m all | nancy sleuth || echo "Nancy not available, skipping security scan"

# Analyze the server-utils replace directive
ls -la ../../pkg/server-utils || echo "server-utils path verification failed"
```

**Success Criteria**:
- All dependencies download successfully
- No version conflicts
- Replace directives point to valid paths
- No known security vulnerabilities

### Phase 1.3: Configuration Validation
**Objective**: Ensure configuration files are valid and complete

**Steps**:
```bash
# Validate YAML syntax
for config in config/*.yaml; do
    echo "Validating $config"
    python3 -c "import yaml; yaml.safe_load(open('$config'))" || echo "YAML validation failed for $config"
done

# Check for required configuration keys
grep -r "cfg\." internal/ | grep -v "test" | sort | uniq > config-usage.log
```

**Success Criteria**:
- All YAML files have valid syntax
- No missing required configuration keys
- Environment-specific configs are consistent

---

## Phase 2: Static Code Analysis

### Phase 2.1: Code Structure Analysis
**Objective**: Identify structural issues and anti-patterns

**Steps**:
```bash
# Run go vet for basic issues
go vet ./...

# Check for unused code
go run honnef.co/go/tools/cmd/staticcheck@latest ./... || echo "staticcheck not available"

# Look for TODO/FIXME/HACK comments
grep -r "TODO\|FIXME\|HACK\|BUG" --include="*.go" .

# Check for hardcoded values
grep -r "http://\|https://\|localhost\|127.0.0.1" --include="*.go" . || true
```

**Bug Categories to Check**:
- Unused variables/imports
- Unreachable code
- Hardcoded credentials/URLs
- Missing error handling
- Potential nil pointer dereferences

### Phase 2.2: Concurrency Analysis
**Objective**: Identify race conditions and concurrency issues

**Steps**:
```bash
# Build with race detector
go build -race -o main-race .

# Check for goroutine management patterns
grep -r "go func\|go " --include="*.go" . > goroutine-analysis.log

# Look for shared state without protection
grep -r "map\[.*\].*{" --include="*.go" . | grep -v "sync\|mutex" > potential-races.log || true

# Check for channel operations
grep -r "make(chan\|<-\|chan " --include="*.go" . > channel-analysis.log
```

**Bug Categories to Check**:
- Race conditions on shared data
- Unbuffered channels causing deadlocks  
- Goroutine leaks
- Missing context cancellation
- Improper channel closing

### Phase 2.3: Error Handling Analysis
**Objective**: Verify comprehensive error handling

**Steps**:
```bash
# Find all error return patterns
grep -n "return.*err\|return.*error" --include="*.go" -r . > error-returns.log

# Check for ignored errors (dangerous pattern)
grep -n "_ =" --include="*.go" -r . | grep -v "test" > ignored-errors.log || true

# Look for panic/recover usage
grep -n "panic\|recover" --include="*.go" -r . > panic-usage.log || true

# Check defer statement usage
grep -n "defer " --include="*.go" -r . > defer-analysis.log
```

**Bug Categories to Check**:
- Unhandled errors
- Incorrect error propagation
- Missing resource cleanup
- Inappropriate panic usage

---

## Phase 3: WebSocket Protocol Analysis

### Phase 3.1: Connection Lifecycle Verification
**Objective**: Verify WebSocket connection handling is correct

**Steps**:
```bash
# Analyze websocket handler implementation
echo "=== WebSocket Handler Analysis ===" > websocket-analysis.log
grep -n "websocket\|conn\|Close\|Write\|Read" internal/handlers/websocket_handler.go >> websocket-analysis.log

# Check connection cleanup patterns
grep -n "defer.*Close\|defer.*close" --include="*.go" -r . >> websocket-analysis.log

# Look for connection state management
grep -n "connections\|clients\|subscriptions" --include="*.go" -r . >> websocket-analysis.log
```

**Bug Categories to Check**:
- Connection leaks
- Improper connection cleanup
- Missing connection state tracking
- Concurrent access to connection objects

### Phase 3.2: Message Broadcasting Analysis
**Objective**: Verify message distribution logic

**Steps**:
```bash
# Analyze broadcasting patterns
grep -n "broadcast\|send\|publish" --include="*.go" -r . > broadcast-analysis.log

# Check subscription management
grep -n "subscribe\|unsubscribe" --include="*.go" -r . > subscription-analysis.log

# Look for message queuing/buffering
grep -n "queue\|buffer\|chan.*Message\|chan.*string" --include="*.go" -r . > messaging-analysis.log || true
```

**Bug Categories to Check**:
- Messages sent to closed connections
- Subscription state inconsistencies
- Message delivery failures
- Buffer overflow conditions

---

## Phase 4: Integration Testing

### Phase 4.1: Local Service Testing
**Objective**: Verify service starts and responds correctly

**Steps**:
```bash
# Start service in background for testing
make run &
SERVICE_PID=$!
echo "Started service with PID: $SERVICE_PID"

# Wait for startup
sleep 5

# Test health endpoint
curl -f http://localhost:8080/health || echo "Health check failed"

# Test metrics endpoint (if enabled)
curl -f http://localhost:8080/metrics || echo "Metrics endpoint not available or failed"

# Test WebSocket endpoint basic connectivity
echo "Testing WebSocket connectivity..." 
timeout 10s wscat -c ws://localhost:8080/ws --execute "echo '{\"type\":\"ping\"}'" || echo "WebSocket test failed (wscat may not be installed)"

# Cleanup
kill $SERVICE_PID 2>/dev/null || true
wait $SERVICE_PID 2>/dev/null || true
```

**Bug Categories to Check**:
- Service startup failures
- Port binding issues
- Endpoint unavailability
- Configuration loading errors

### Phase 4.2: Delta Exchange Integration Testing
**Objective**: Verify external service integration

**Steps**:
```bash
# Check Delta Exchange connection logic
echo "=== Delta Exchange Integration Analysis ===" > delta-integration.log
grep -n "delta\|Delta\|websocket.*connect" --include="*.go" -r . >> delta-integration.log

# Look for authentication handling
grep -n "auth\|token\|key\|secret" --include="*.go" -r . >> delta-integration.log

# Check retry/reconnection logic
grep -n "retry\|reconnect\|backoff" --include="*.go" -r . >> delta-integration.log || true
```

**Bug Categories to Check**:
- Authentication failures
- Connection timeout issues
- Missing retry logic
- Improper error handling on disconnection

---

## Phase 5: Stress & Load Testing

### Phase 5.1: Concurrent Connection Testing
**Objective**: Test behavior under multiple concurrent connections

**Steps**:
```bash
# Start service for load testing
make run &
SERVICE_PID=$!
sleep 5

# Create multiple WebSocket connections (if tools available)
echo "=== Load Testing Setup ===" > load-test.log
echo "Service PID: $SERVICE_PID" >> load-test.log

# Manual concurrent connection simulation
for i in {1..5}; do
    (curl -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" http://localhost:8080/ws &) 2>/dev/null
done

sleep 10

# Check for resource leaks
ps aux | grep -E "(main|websocket)" >> load-test.log
netstat -an | grep ":8080\|:9090" >> load-test.log || ss -tuln | grep ":8080\|:9090" >> load-test.log

# Cleanup
kill $SERVICE_PID 2>/dev/null || true
wait $SERVICE_PID 2>/dev/null || true
killall main 2>/dev/null || true
```

**Bug Categories to Check**:
- Memory leaks under load
- Connection limit issues  
- Resource exhaustion
- Performance degradation

### Phase 5.2: Message Throughput Testing
**Objective**: Test high-frequency message handling

**Steps**:
```bash
# Test message handling under load
echo "=== Message Throughput Analysis ===" > throughput-test.log

# Analyze message processing paths
grep -n "message\|Message" --include="*.go" -r . | grep -v "test" >> throughput-test.log

# Look for potential bottlenecks
grep -n "mutex\|lock\|sync\|channel" --include="*.go" -r . >> throughput-test.log
```

**Bug Categories to Check**:
- Message processing bottlenecks
- Channel blocking issues
- Lock contention problems
- Buffer overflow conditions

---

## Phase 6: Security Analysis

### Phase 6.1: Input Validation Testing
**Objective**: Verify proper input sanitization and validation

**Steps**:
```bash
# Check for input validation patterns
echo "=== Security Analysis ===" > security-analysis.log
grep -n "json\|JSON\|Unmarshal\|decode" --include="*.go" -r . >> security-analysis.log

# Look for potential injection points
grep -n "string.*format\|fmt\.Sprintf" --include="*.go" -r . >> security-analysis.log || true

# Check authentication implementation
grep -n -i "auth\|token\|secret\|password" --include="*.go" -r . >> security-analysis.log
```

**Bug Categories to Check**:
- JSON injection vulnerabilities
- Missing input validation
- Authentication bypasses
- Information disclosure

### Phase 6.2: Resource Protection Analysis
**Objective**: Verify protection against resource abuse

**Steps**:
```bash
# Check for rate limiting
grep -n -i "rate\|limit\|throttle" --include="*.go" -r . > rate-limiting.log || echo "No rate limiting found" > rate-limiting.log

# Look for resource cleanup
grep -n "defer\|Close\|cleanup" --include="*.go" -r . > resource-cleanup.log

# Check timeout implementations
grep -n -i "timeout\|deadline\|context" --include="*.go" -r . > timeout-analysis.log
```

**Bug Categories to Check**:
- Missing rate limiting
- Resource exhaustion vulnerabilities
- Timeout handling issues
- Improper resource cleanup

---

## Phase 7: Final Integration & Regression Testing

### Phase 7.1: End-to-End Workflow Testing
**Objective**: Test complete user workflows

**Steps**:
```bash
# Test complete subscription workflow
echo "=== End-to-End Testing ===" > e2e-test.log

# Start service
make run &
SERVICE_PID=$!
sleep 5

# Record the test
echo "Testing complete WebSocket workflow..." >> e2e-test.log
echo "1. Connect to WebSocket" >> e2e-test.log
echo "2. Send subscription message" >> e2e-test.log  
echo "3. Verify message receipt" >> e2e-test.log
echo "4. Send unsubscribe message" >> e2e-test.log
echo "5. Verify cleanup" >> e2e-test.log

# Cleanup
kill $SERVICE_PID 2>/dev/null || true
wait $SERVICE_PID 2>/dev/null || true
```

**Bug Categories to Check**:
- Workflow interruption bugs
- State inconsistency issues
- Integration failures
- User experience problems

### Phase 7.2: Regression Testing
**Objective**: Ensure no existing functionality is broken

**Steps**:
```bash
# Run any existing tests
go test ./... -v > regression-test.log 2>&1 || echo "Some tests failed"

# Check test coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html 2>/dev/null || echo "Coverage report generation skipped"

# Verify all endpoints still work
echo "=== Regression Testing Summary ===" >> regression-test.log
echo "All previously working functionality should remain operational" >> regression-test.log
```

**Bug Categories to Check**:
- Regression in existing features
- Test failures
- Coverage gaps
- Integration breakages

---

## Phase Completion Protocol

After each phase:

1. **Generate Phase Summary**: Create `phase-N-summary.md` with:
   - Phase completion status
   - Number of issues found  
   - Critical findings summary
   - Next phase readiness

2. **Update Bug Index**: If bugs found, update `docs/bugs/00-overview_of_bugs.md`

3. **Clean Temporary Files**: Remove analysis logs and temporary files

4. **Commit Progress**: Commit documentation updates with descriptive messages

---

## Success Metrics

| Metric | Target | Critical Threshold |
|--------|---------|-------------------|
| Build Success Rate | 100% | Cannot proceed if build fails |
| Critical Bugs Found | Document all | > 0 requires immediate attention |
| Test Coverage | > 70% | < 50% indicates insufficient testing |
| Performance Issues | 0 under normal load | Any degradation needs investigation |
| Security Vulnerabilities | 0 | Any found requires immediate fix |

---

## Tools Required

| Tool | Purpose | Installation |
|------|---------|-------------|
| `go` | Build and test | Required - Go 1.21+ |
| `make` | Build automation | Usually pre-installed |
| `curl` | HTTP endpoint testing | Usually pre-installed |
| `wscat` | WebSocket testing | `npm install -g wscat` |
| `staticcheck` | Advanced static analysis | `go install honnef.co/go/tools/cmd/staticcheck@latest` |
| `nancy` | Security scanning | `go install github.com/sonatypecommunity/nancy@latest` |

---

## Documentation Updates

This playbook should be updated when:
- New phases are added
- Tool requirements change
- Bug patterns are discovered
- Process improvements are identified

**Last Updated**: Initial Version  
**Next Review**: After first complete execution 