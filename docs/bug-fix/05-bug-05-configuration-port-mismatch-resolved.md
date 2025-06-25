# Configuration Port Mismatch - Medium

**Bug ID**: 05-bug-05
**Discovery Phase**: Phase 3.1
**Severity**: Medium
**Status**: Resolved
**Reporter**: Bug Identification Process
**Date Discovered**: 2024-06-24

---

## Summary
The service ignored port values from `local.yaml` and always started on the default ports. Logging and environment override handling were missing making the issue hard to diagnose.

## Code Changes
- `internal/config/config.go` – added logging and support for `HTTP_PORT` and `GRPC_PORT` overrides
- `tests/bugs/05_repro_test.go` – reproduction test ensuring ports from config are honored
- `docs/bugs/05-bug-05-configuration-port-mismatch.md` – marked as Fixed

## Tests Added
- `tests/bugs/05_repro_test.go`

## Verification
```bash
# on dev branch the test fails
# on bug/05-configuration-port-mismatch it passes
make test-race
```

## Checklist
- [x] `make ci` passes
- [x] Bug status updated
- [x] Resolved report added
- [x] Failing test reproduced and now passes
