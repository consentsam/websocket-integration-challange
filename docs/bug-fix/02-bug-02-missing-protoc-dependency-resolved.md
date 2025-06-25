# 02-bug-02 Missing Protocol Buffer Compiler Dependency - Resolved

The build process failed when `protoc` was not installed. A check was added to the `Makefile` and installation instructions were documented.

## Verification

1. Temporarily remove `protoc` from `PATH`:
   ```bash
   PATH=/usr/bin:/bin make proto
   ```
   The command should output `Error: protoc not found` with installation hints and exit with a non-zero status.

2. Install `protoc` using the provided instructions and rerun `make proto`. It should succeed.
