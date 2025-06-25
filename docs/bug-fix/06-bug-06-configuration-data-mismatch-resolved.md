# Bug Fix Report: 06-bug-06 Configuration Data Mismatch in Product IDs

## Summary
The configuration loader failed to locate YAML files when the service binary was executed from a different directory. As a result, default values (e.g. `BTCUSD`) were used instead of those specified in `config/local.yaml` such as `BTC_USDT`.

## Resolution
`LoadConfig` now searches for configuration files relative to the executable and the source tree, ensuring that the correct YAML file is found regardless of the working directory. Additional logging of the loaded Delta `ProductIDs` helps verify the final configuration.

## Verification
1. Run `go test -race ./tests/bugs -run TestBug06_Repro -v`.
   - Before the fix: test failed showing `BTCUSD`.
   - After the fix: test passes and logs `BTC_USDT`.
2. Execute `make ci` to ensure all checks pass.
