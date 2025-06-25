# Bug 12 Resolved - Unsubscribe Client Count Logic Error

The handler incorrectly checked channel subscriber counts before removing the client. This prevented unsubscribing from Delta Exchange when the last client left a channel.

## Fix Summary
- `handleUnsubscribe` now removes the client first and then determines if any subscribers remain.
- Added unit test `TestBug12_Repro` validating the Delta client only unsubscribes when the last client disconnects.

## Verification
1. `make ci` passes.
2. Running `go test ./...` shows `TestBug12_Repro` succeeds.
