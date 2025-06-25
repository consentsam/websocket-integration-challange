# 10-bug-10 Read Lock Write Operation Race Condition - Resolved

The broadcast loop attempted to delete clients from the connection map while only holding a read lock. This violated mutex semantics and triggered race warnings under load.

## Verification
1. Added regression test `TestBug10_Repro` to simulate a blocked client and ensure broadcasts remove it safely.
2. Updated `run()` broadcast logic to gather failed clients under `RLock` and delete them under `Lock`.
3. Ran `make ci` to confirm build, lint, and tests pass with race detector.

## Result
Broadcasting now cleans up stalled clients without data races or panics.
