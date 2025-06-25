# Bug 11 Resolved - Malformed JSON Message Batching

**Bug ID**: 11-bug-11
**Discovery Phase**: Phase 2.1
**Severity**: High
**Status**: Resolved
**Reporter**: Phase 2 Verification Analysis
**Date Discovered**: 2024-12-19

---

## Summary
The `writePump` function concatenated queued JSON messages using newline
characters, resulting in invalid JSON for clients. The handler now writes each
queued message as a separate WebSocket frame.

## Code Changes
- `internal/handlers/websocket_handler.go` – send queued messages separately
  instead of newline batching.
- `tests/bugs/11_repro_test.go` – regression test verifying valid JSON frames.

## Verification
```bash
make test -race ./tests/bugs -run TestBug11_Repro
```

## Checklist
- [x] Updated bug status in documentation
- [x] Added regression test
- [x] `make ci` passes
- [x] Created this resolved bug report
