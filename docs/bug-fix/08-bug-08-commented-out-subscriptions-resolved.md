# 08-bug-08 Commented Out Channel Subscriptions - Resolved

The Delta WebSocket client was connecting without subscribing to any channels because the subscription loop in `Connect()` was commented out. This prevented market data from being received automatically.

## Verification
1. Added a regression test `TestBug08_Repro` which starts a local WebSocket server and ensures a subscribe message is sent after `Connect()`.
2. Uncommented the subscription logic in `internal/clients/delta_websocket.go`.
3. Ran `make ci` to confirm build, lint, and tests all pass.

## Result
`Connect()` now automatically subscribes to configured channels and the regression test passes.
