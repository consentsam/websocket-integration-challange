# Race Condition in WebSocket Handler Broadcast - High

**Bug ID**: 03-bug-03
**Discovery Phase**: Phase 2.2
**Severity**: High
**Status**: Resolved
**Reporter**: Bug Identification Process
**Date Discovered**: 2024-06-24

---

## Summary
The broadcast logic in `internal/handlers/websocket_handler.go` modified the
`clients` map while only holding a read lock. When other goroutines accessed the
handler at the same time (e.g. `GetStatistics`), the race detector reported a
concurrent map access error. The fix collects failed clients under the read lock
and performs map deletions with a write lock.

## Code Changes
- `internal/handlers/websocket_handler.go` – two-phase broadcast cleanup to avoid
  map writes under `RLock`.
- `internal/handlers/websocket_handler_test_helpers.go` – test helper functions
  (build tag `test`).
- `tests/bugs/03_repro_test.go` – reproduction test demonstrating the race.

## Tests Added
- `tests/bugs/03_repro_test.go`

## Verification
```bash
# Reproduce failure on dev branch (expected race)
make test -race ./tests/bugs -run TestBug03_Repro

# On bug branch (race free)
make test -race ./tests/bugs -run TestBug03_Repro
```

## Checklist
- [x] Updated bug status in documentation
- [x] Added regression test
- [x] `make ci` passes
- [x] Created this resolved bug report
