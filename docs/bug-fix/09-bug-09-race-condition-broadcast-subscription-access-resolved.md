# 09-bug-09 Race Condition in BroadcastToChannel Subscription Access - Resolved

Concurrent modifications to the subscription map could occur while broadcasting
messages. `BroadcastToChannel` copied the reference to the clients map while
holding only a read lock, then iterated without any synchronization. This led to
concurrent map writes and panics under the race detector.

## Verification
1. Added regression test `TestBug09_Repro` that runs broadcasts and unsubscriptions
   concurrently. Prior to the fix, the race detector reports a data race.
2. Updated `BroadcastToChannel` to copy the subscribed clients while the
   subscription lock is held and verify membership before sending.
3. `make ci` confirms all tests pass with the race detector enabled.

## Result
Broadcasting to a channel no longer races with subscription updates and tests pass
cleanly under `-race`.

