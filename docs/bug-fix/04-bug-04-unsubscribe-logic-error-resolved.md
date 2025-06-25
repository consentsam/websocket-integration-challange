# Bug 04 Resolved - Logic Error in Unsubscribe Client Count Check

The unsubscribe logic checked the number of subscribed clients *before* removing the current client. As a result, the last client leaving a channel never triggered a Delta Exchange unsubscription.

## Fix Summary
- `handleUnsubscribe` now removes the client first and then checks the remaining client count under a read lock.
- Introduced `clients.DeltaClient` interface to allow testing with a stub Delta client.
- Added unit test `TestBug04_Repro` reproducing the issue and verifying the fix.

## Verification
1. `make ci` completes successfully.
2. Running `go test ./...` shows `TestBug04_Repro` passes.
