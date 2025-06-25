# AGENTS.md

A single, authoritative operating manual for **humans and AI agents (e.g. CODEX in ChatGPT)** who contribute to this repository.

---
## 1. Mission
Fix, test and merge every documented bug (`docs/bugs/**/*.md` and `docs/bugs-phase-02/**/*.md`).  Each bug is resolved in its own branch and verified by automated CI.

---
## 2. Repository Rules
1. **Never** commit directly to `main` or `dev-*` branches.
2. Each bug ⇒ dedicated feature branch: `bug/<bug-id>-<kebab-title>`  
   Example: `bug/03-race-condition-websocket-handler`.
3. Keep documentation in sync with code (update bug index, status fields, counters).
4. Create a resolved-bug report under `docs/bug-fix/` once the fix is merged.  
   File name format: `<bug-id>-<kebab-title>-resolved.md` (e.g. `01-bug-01-missing-proto-generation-resolved.md`).
5. **Trigger phrase**: asking `RESOLVE THE BUG <bug-file>` kicks off the automated fix workflow.
5. Squash-merge every PR via GitHub UI.

---
## 3. Development Workflow (TL;DR)
```bash
# 0. Pre-requisites (first time only)
make deps proto            # download modules & generate code
make ci                    # run local mirror of CI pipeline (works offline)

# 1. Pick a bug & create branch
BUG=03-bug-03-race-condition-websocket-handler
BRANCH="bug/${BUG%%-*}-${BUG#*-}"
git checkout -b "$BRANCH"

# 2. Reproduce failure with a new test (should RED)
#    ‑ add *_test.go reproducer per bug report instructions
make test -race | cat

# 3. Implement the fix & update docs
#    ‑ follow Recommended Fix in markdown file
#    ‑ update Status in bug file → Fixed
#    ‑ run `make proto` if proto files changed
#    ‑ create resolved report in docs/bug-fix/<bug-id>-<title>-resolved.md

# 4. Verify locally
make ci                    # build, vet, lint, tests, race, proto-diff (works offline)

# 5. Commit & push
git add .
git commit -s -m "fix(websocket): race in broadcast map (03-bug-03)"
git push --set-upstream origin "$BRANCH"

# 6. Open PR (template auto-loads)
#    PR title:  "fix: 03-bug-03 – race in websocket broadcast"
```

Detailed per-bug steps are documented in section 8.

---
## 4. Pull-Request Checklist (enforced)
- [ ] Links to bug markdown (e.g. `closes docs/bugs/03-bug-03-race-condition-websocket-handler.md`).
- [ ] `make ci` passes (build + vet + staticcheck + tests ‑race + proto diff clean) — works offline.
- [ ] New or updated tests cover the fix (failing on `dev` / passing on branch).
- [ ] Documentation updated (bug Status, overview counters, CHANGELOG if needed).
- [ ] Resolved report added under `docs/bug-fix/` with verification steps.
- [ ] At least one reviewer approval.

---
## 5. Coding Standards
* Go **1.22**.  Run `goimports` on save.
* **Concurrency**: no map writes under `RLock`; always run `go test -race`.
* Return `%w` wrapped errors; log using `zap` (`debug|info|warn|error`).
* No TODOs committed to `main`.

---
## 6. Testing Standards
* All fixes require a failing test first (Red-Green-Refactor).
* Race detector and coverage (≥70 %) run in CI.
* Integration tests live under `tests/`.

---
## 7. Branch & Commit Conventions
| Item   | Format | Example |
|--------|--------|---------|
| **Branch** | `bug/<id>-<kebab-title>` | `bug/05-configuration-port-mismatch` |
| **Commit** | `fix(<area>): <short> (#bug-id)` | `fix(websocket): unsubscribe logic error (#12)` |
| **PR title** | `fix: <bug-id> <short>` | `fix: 09-bug-09 race in broadcast` |

---
## 8. Per-Bug Operating Procedure (expanded)
1. **Branch** – `git checkout -b bug/<id>-<title>`.
2. **Test First** – translate reproduction steps into failing test.
3. **Verify RED** – `make test -race`; ensure failure.
4. **Fix** – implement code & docs changes.
5. **Local CI** – `make ci` (works offline).
6. **Docs** – mark bug as `Fixed`, update overview counts.
7. **Push & PR** – PR auto-fills template; complete checklist.
8. **Review & Merge** – squash merge once green.

---
## 9. Tooling Commands (quick reference)
| Purpose      | Command                     |
|--------------|----------------------------|
| Install deps | `make deps`                |
| Generate PB  | `make proto`               |
| Build app    | `make build`               |
| Full CI      | `make ci` (offline ready)  |
| Hot reload   | `make dev`                 |
| Coverage     | `make test-coverage`       |

---
## 10. Automation Matrix
| Stage  | Trigger | Agent / Tool               |
|--------|---------|----------------------------|
| CI     | push/PR | GitHub Actions (`ci.yml`)  |
| Review | PR      | Human / ChatGPT-CODEX      |
| Merge  | PR green & approval | GitHub UI |

---
## 11. Fallback / Escalation
If CI fails with a non-trivial race or flaky test, label the PR `needs-investigation` in a comment. 

---
## 12. Bug-Fix Documentation (`docs/bug-fix/`)

Every resolved bug **must** include a standalone markdown report located at `docs/bug-fix/<bug-id>-<kebab-title>-resolved.md`.

**Required sections:**

| Section | Purpose |
|---------|---------|
| Header  | Copy original bug header (ID, Severity, Date) + add `Status: Resolved` |
| Summary | Short description of root cause & fix approach |
| Code Changes | Bullet list of affected files / functions |
| Tests Added | Paths to new/updated `*_test.go` files |
| Verification | CLI commands to reproduce failure on `dev` & success on branch |
| Checklist | Tick-box PR checklist (mirrors Section 4) |

Include any diagrams, logs, or links that help reviewers understand the fix.  Place binary assets (e.g. PNG) in `docs/bug-fix/assets/`.

---
### Deep-Wiki & API Docs locations

* Internal architecture docs live under `docs/deepwiki-docs/`
* External API reference snapshots live under `docs/delta-exchange-api-docs/`

Refer to these paths when cross-linking from resolved reports. 

---
## 13. Telemetry Logging Implementation Workflow

The observability roadmap lives in `docs/telemetry-planning.md`.  Each **phase** (P-1 … P-7) is delivered with the same discipline used for bug fixes, but through a dedicated trigger phrase so that CODEX can automate scaffolding and quality-gates.

### 13.1 Trigger Phrase
> **IMPLEMENT THE PHASE-<n> OF LOGGING in docs/telemetry-planning.md**

Using the phrase above (replace `<n>` with 1-7) in a PR description or chat message immediately spins up the following automated workflow:
1. Create a feature branch `telemetry/phase-<n>`.
2. Apply code changes listed for that phase in `docs/telemetry-planning.md`.
3. Run `make ci` (works offline) and telemetry smoke-tests.
4. Open a draft PR ready for review.

### 13.2 Per-Phase Checklist (CI Enforced)
- [ ] Code matches exact insertion points & items in `docs/telemetry-planning.md`.
- [ ] All existing tests + new telemetry tests pass (`make ci` — works offline).
- [ ] `/metrics` endpoint returns **HTTP 200** and exposes new metrics family.
- [ ] Jaeger (or Tempo) shows at least one span emitted by this phase (smoke test).
- [ ] Feature flag `TELEMETRY_PHASE_<n>_ENABLED` defaults **true** in dev, **false** in prod charts until reviewed.
- [ ] Documentation — toggle corresponding checkbox in Section 6 of `telemetry-planning.md`.

### 13.3 Branch & Commit Convention
| Item   | Format                                    | Example                               |
|--------|-------------------------------------------|---------------------------------------|
| Branch | `telemetry/phase-<n>-<short-title>`       | `telemetry/phase-3-broadcast-hub`     |
| Commit | `telemetry(<area>): phase-<n> <summary>`  | `telemetry(websocket): phase-2 ingest`|
| PR     | `telemetry: phase-<n> – <short-title>`    | `telemetry: phase-1 core bootstrap`   |

Follow the existing review, merge and escalation rules (Sections 4, 10 & 11) for all telemetry work. 

---
## 14. Offline Builds & CODEX Compatibility

Since CODEX (ChatGPT) environments don't have internet access, the repository includes a `vendor/` directory with all Go dependencies for offline builds.

### 14.1 How It Works
- **With internet** (local development): `make ci` downloads dependencies as needed and runs full CI pipeline
- **Without internet** (CODEX): `make ci` automatically uses vendored dependencies with `-mod=vendor` and runs the same full CI pipeline

### 14.2 Maintaining Vendor Directory
When dependencies change, update the vendor directory:
```bash
make vendor
git add vendor/ go.mod go.sum
git commit -m "vendor: update dependencies for CODEX compatibility"
```

### 14.3 Makefile Auto-Detection & Protobuf Compatibility
The `Makefile` automatically detects if `vendor/` exists and switches to offline mode:
- ✅ **Offline mode**: Uses `-mod=vendor` flags when `vendor/` directory exists
- 📥 **Online mode**: Downloads dependencies when `vendor/` directory is missing

**Protobuf Version Compatibility:**
- `proto-check` target ignores protoc version differences in generated file comments
- Local dev (protoc v5.29.3) vs CODEX (protoc v5.27.1) differences are automatically ignored
- Only actual code changes in protobuf files will cause CI failures

This ensures both local development and CODEX environments work seamlessly.

---
## 15. Verification Workflow for CODEX/ChatGPT Environment

When working with CODEX in ChatGPT by OpenAI, follow this verification workflow to ensure changes work correctly in the offline environment.

### 15.1 Pre-Verification Setup
```bash
# Ensure offline builds work
make ci                    # Full CI pipeline (works offline)
make build                 # Build the service binary
```

### 15.2 Telemetry Verification Workflow

#### **Step 1: Service Startup Verification**
```bash
# Test service starts without errors
timeout 3 ./websocket-service 2>&1 | head -10

# Expected output should show:
# - Configuration loading from YAML
# - Metrics endpoint enabled
# - No panic or startup errors
```

#### **Step 2: Metrics Endpoint Verification**
```bash
# Start service in background
./websocket-service &
SERVICE_PID=$!

# Wait for startup
sleep 2

# Verify metrics endpoint responds
curl -f http://localhost:8080/metrics > /dev/null && echo "✅ Metrics OK" || echo "❌ Metrics FAIL"

# Check specific metrics are present
curl -s http://localhost:8080/metrics | grep -E "delta_messages_total|broadcast_total|go_panic_total" | head -5

# Clean up
kill $SERVICE_PID 2>/dev/null || true
```

#### **Step 3: Configuration Verification**  
```bash
# Test different environments
ENVIRONMENT=development timeout 3 ./websocket-service 2>&1 | grep -E "config|port|metrics"

# Test environment variable overrides
WEBSOCKET_HTTP_PORT=9999 timeout 3 ./websocket-service 2>&1 | grep "HTTP Port: 9999"

# Test metrics disable
WEBSOCKET_METRICS_ENABLED=false timeout 3 ./websocket-service 2>&1 | grep "Metrics Enabled: false"
```

### 15.3 Code Change Verification

#### **For Bug Fixes:**
```bash
# 1. Verify fix compiles
make build

# 2. Run specific tests (if race-related)
go test -race ./internal/handlers/ -v

# 3. Run full test suite
make test

# 4. Verify no new lint issues
make lint
```

#### **For Telemetry Changes:**
```bash
# 1. Verify no build errors
go build -o websocket-service .

# 2. Check telemetry initialization
./websocket-service 2>&1 | head -5 | grep -E "telemetry|metrics"

# 3. Validate metrics schema
curl -s http://localhost:8080/metrics | grep "^# HELP" | head -10

# 4. Test panic recovery (if implemented)
# This would be tested through specific test cases
```

### 15.4 Integration Testing in CODEX

#### **WebSocket Functionality Test:**
```bash
# Start service
./websocket-service &
SERVICE_PID=$!
sleep 2

# Test WebSocket endpoint accessibility (connection test)
# Note: Full WebSocket testing requires external tools not available in CODEX
# But we can verify the service accepts connections on the right port
netstat -ln | grep :8080 || ss -ln | grep :8080

# Verify gRPC port
netstat -ln | grep :9090 || ss -ln | grep :9090

kill $SERVICE_PID 2>/dev/null || true
```

#### **Error Handling Verification:**
```bash
# Test service handles invalid config gracefully
echo "invalid: yaml: content" > /tmp/invalid.yaml
WEBSOCKET_CONFIG_FILE=/tmp/invalid.yaml timeout 3 ./websocket-service 2>&1 | grep -i error

# Test service handles missing config
ENVIRONMENT=nonexistent timeout 3 ./websocket-service 2>&1
```

### 15.5 Documentation Verification

#### **Verify Documentation Changes:**
```bash
# Check markdown syntax
# (In CODEX, visual inspection of markdown is needed)

# Verify file references exist
ls -la docs/telemetry-implementation.md
ls -la docs/telemetry-quick-reference.md

# Check that documentation examples work
grep -A 5 "curl.*metrics" docs/telemetry-quick-reference.md
```

### 15.6 Common CODEX Verification Patterns

#### **Quick Health Check:**
```bash
# One-liner service health verification
./websocket-service & sleep 2 && curl -f http://localhost:8080/metrics >/dev/null && echo "✅ Service + Metrics OK" || echo "❌ Failed"; pkill -f websocket-service 2>/dev/null
```

#### **Configuration Matrix Test:**
```bash
# Test all environment configs exist and load
for env in local development production; do
  echo "Testing $env..."
  ENVIRONMENT=$env timeout 2 ./websocket-service 2>&1 | grep -E "config|port" | head -3
  echo "---"
done
```

#### **Metrics Completeness Check:**
```bash
# Verify all expected metrics are exported
./websocket-service & sleep 2
EXPECTED_METRICS="delta_messages_total json_unmarshal_errors_total broadcast_total broadcast_latency_ms client_delivery_total go_panic_total"
for metric in $EXPECTED_METRICS; do
  curl -s http://localhost:8080/metrics | grep -q "$metric" && echo "✅ $metric" || echo "❌ $metric MISSING"
done
pkill -f websocket-service 2>/dev/null
```

### 15.7 Troubleshooting CODEX Issues

#### **Common Problems & Solutions:**

| Problem | Symptom | Solution |
|---------|---------|----------|
| **Build fails** | `go: module not found` | Run `make deps` or `go mod vendor` |
| **Service won't start** | `panic: invalid pattern` | Check configuration loading in logs |
| **Metrics empty** | `/metrics` returns 200 but no data | Verify Prometheus exporter setup |
| **Port conflicts** | `bind: address already in use` | Use different ports or kill existing processes |
| **Config not loading** | Uses wrong values | Check YAML file paths and env vars |

#### **Debug Commands:**
```bash
# Check what ports are in use
netstat -tlnp 2>/dev/null | grep -E ":808[0-9]|:909[0-9]" || ss -tlnp | grep -E ":808[0-9]|:909[0-9]"

# Verify service binary
file ./websocket-service
ldd ./websocket-service 2>/dev/null || otool -L ./websocket-service 2>/dev/null || echo "Static binary"

# Check configuration file loading
strace -e openat ./websocket-service 2>&1 | grep -E "\.yaml" | head -5 2>/dev/null || \
dtruss -n ./websocket-service 2>&1 | grep -E "\.yaml" | head -5 2>/dev/null || \
echo "Config loading verification requires system tracing tools"
```

### 15.8 Final Verification Checklist for CODEX

Before considering any change complete in CODEX environment:

- [ ] **Build**: `make ci` passes without internet connectivity
- [ ] **Startup**: Service starts without errors and logs configuration
- [ ] **Endpoints**: HTTP/gRPC ports are accessible
- [ ] **Metrics**: `/metrics` endpoint returns valid Prometheus data
- [ ] **Config**: Environment variables and YAML files load correctly
- [ ] **Tests**: Relevant test cases pass with `-race` flag
- [ ] **Documentation**: Examples in docs actually work when executed
- [ ] **Cleanup**: No orphaned processes or temp files left behind

This verification workflow ensures that changes made through CODEX in ChatGPT are production-ready and function correctly in offline environments.

---
## 16. Merge-Order Playbook

When multiple bug-fix branches touch overlapping files, follow this playbook to avoid merge hell and preserve a linear history.

### 16.1 Golden Rules
1. **Rebase, don't merge**: Always `git rebase origin/main` before opening (or refreshing) a PR.  
   *Why?* Keeps history clean and prevents multi-merge diamonds.
2. **Resolve conflicts locally** (never in the GitHub UI). Use:
```bash
git rebase origin/main               # interactive if needed
git mergetool                        # or resolve by hand
make ci                               # ensure build still green
git push --force-with-lease           # update PR safely
```
3. **Enable rerere** once per workstation so Git remembers conflict resolutions:
```bash
git config --global rerere.enabled true
```
4. **Generated code in a separate commit**: If you regenerate protobufs or mocks, push them in a follow-up commit (`chore(proto): regenerate`) so the *functional* changes stay conflict-friendly.

### 16.2 Typical Conflict Scenarios & Fixes
| Scenario | Symptom | Fix |
|----------|---------|-----|
| Two bug branches edit the same Go file | `<<<<<<< HEAD` markers | Rebase latest, accept both fixes, run `goimports` |
| Proto files regenerated in both branches | Large diff in `*_pb.go` | Keep newer proto, re-run `make proto`, commit generated surge |
| Docs conflict in bug overview table | Table row duplicated | Keep both rows, re-sort by Bug-ID |

### 16.3 PR Merge Order Guidance
1. **Critical severity first** (Section 2 labels).  
2. **Smaller diff size wins** if severity equal.  
3. **Conflicting branches**: merge the one that touches *fewer* files first, then the larger one rebases.

---
## 17. Reproduction Test Isolation

Every bug MUST supply a failing _reproduction test_ that demonstrates the issue before the fix.  To avoid test-name clashes and ensure clarity, adopt the following rules.

### 17.1 File & Function Naming
| Item | Rule | Example |
|------|------|---------|
| Test file | `tests/bugs/<bug-id>_repro_test.go` | `tests/bugs/03_repro_test.go` |
| Test func | `TestBug<id>_Repro` | `TestBug03_Repro` |

### 17.2 Lifecycle
1. **Initial commit**: Test fails (RED) on `dev` branch.  
2. **Fix commit**: Test passes (GREEN) on bug branch.  
3. **Post-merge**: Convert test to **skipped mode** so main stays green _but regression is tracked_.
```go
func TestBug03_Repro(t *testing.T) {
    if os.Getenv("CI") == "true" { t.Skip("Regression test; passes post-fix") }
    // original assertions here
}
```

### 17.3 CI Integration
* `make ci` treats any test under `tests/bugs/` that **is not skipped** as a failure.  
* Regression tests run nightly via the `schedule.yml` workflow to ensure no future regressions.

---
## 18. Automated Bug Index & Counters

Manual counting of open vs fixed bugs is error-prone when many issues are tackled in parallel.  A tiny helper tool keeps the index in sync.

### 18.1 Usage
```bash
# Generate a fresh index table & counts
go run ./cmd/bugindex > docs/bugs/00-overview_of_bugs.md
```
The command parses every markdown file under `docs/bugs/**` and `docs/bugs-phase-02/**`, extracts the front-matter fields (`ID`, `Title`, `Severity`, `Status`) and emits an updated overview table with aggregate counters.

### 18.2 Makefile Target
Add once per clone:
```bash
make bug-index        # Wrapper for the go run command above
```

### 18.3 Implementation Notes
* Located at `cmd/bugindex/main.go` (see source).
* Ignores files under `docs/bug-fix/`.
* Treats statuses case-insensitively: `Open|Fixed|In Progress`.
* Fails CI if counts in overview don't match reality (prevents stale docs).

---