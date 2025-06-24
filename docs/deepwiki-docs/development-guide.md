# Development Guide

**Relevant source files**
* [.gitignore](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/.gitignore)
* [Makefile](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile)
* [go.mod](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/go.mod)
* [go.sum](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/go.sum)
* [main.go](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go)

This guide covers building, testing, and developing the websocket-service locally. It includes dependency management, build automation, code generation workflows, and development best practices.

For configuration management across environments, see [Configuration Guide](#5). For production deployment practices, see [Deployment Guide](#7).

## Development Environment Setup

### Prerequisites

The development environment requires the following tools and versions:

| Tool | Version | Purpose |
| :--- | :--- | :--- |
| `go` | 1.22.0+ | Go runtime and compiler |
| `protoc` | Latest | Protocol buffer compiler |
| `docker` | Latest | Container building and running |
| `air` | Latest | Hot reload development server |

### Project Structure

The build system is organized around these key components:

**Development Workflow Dependencies**
Sources: [Makefile#L1-L86](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L1-L86) [go.mod#L1-L23](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/go.mod#L1-L23) [main.go#L1-L117](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L1-L117)

## Build System Overview

The build system uses GNU Make for automation with standardized targets:

### Core Build Targets

| Target | Command | Purpose |
| :--- | :--- | :--- |
| `all` | `make all` | Complete build: clean, deps, proto, build |
| `build` | `make build` | Compile binary with optimization flags |
| `clean` | `make clean` | Remove build artifacts and generated code |
| `deps` | `make deps` | Download and tidy Go dependencies |
| `proto` | `make proto` | Generate Go code from Protocol Buffers |

### Development Targets

| Target | Command | Purpose |
| :--- | :--- | :--- |
| `run` | `make run` | Start service with `go run main.go` |
| `dev` | `make dev` | Hot reload development with `air` |
| `test` | `make test` | Run test suite with verbose output |
| `test-coverage` | `make test-coverage` | Generate HTML coverage report |

**Build Process Flow**

Sources: [Makefile#L18-L44](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L18-L44) [Makefile#L55-L68](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L55-L68)

## Dependency Management

### Go Module Configuration

The service uses Go modules for dependency management with the following configuration:

**Dependency Commands**

| Command | Purpose |
| :--- | :--- |
| `go mod download` | Download dependencies to module cache |
| `go mod tidy` | Add missing and remove unused dependencies |
| `go mod verify` | Verify dependencies have expected content |
| `go mod graph` | Print module requirement graph |

Sources: [go.mod#L1-L23](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/go.mod#L1-L23) [go.sum#L1-L37](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/go.sum#L1-L37) [Makefile#L35-L38](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L35-L38)

## Code Generation

### Protocol Buffer Workflow

The service generates Go code from Protocol Buffer definitions:

**Generated Code Integration**

The generated protobuf code integrates with the main application through these imports:

* [main.go#L14-L14](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L14-L14): `websocketv1 "github.com/Cryptovate-India/websocket-service/gen/websocket/api/v1"`
* [main.go#L53-L53](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L53-L53): `websocketv1.RegisterWebsocketServiceServer(grpcServer, websocketServer)`

Sources: [Makefile#L40-L44](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L40-L44) [main.go#L14-L14](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L14-L14) [main.go#L53-L53](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L53-L53)

## Development Workflow

### Local Development Process

### Hot Reload Development

For rapid development iteration, use the hot reload server:

```shellscript
make dev
````

This command uses `air` with configuration from `.air.toml` to:

  * Watch source file changes
  * Automatically rebuild and restart the service
  * Preserve development state during restarts

### Testing Workflow

| Test Type | Command | Output |
| :--- | :--- | :--- |
| Unit Tests | `make test` | Verbose test output to console |
| Coverage Report | `make test-coverage` | HTML report in `coverage.html` |
| Manual Testing | `make run` | Service available on ports 8080/9090 |

Sources: [Makefile\#L22-L24](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L22-L24) [Makefile\#L66-L68](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L66-L68) [Makefile\#L56-L63](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L56-L63)

## Application Bootstrap

### Startup Sequence

The main application follows this initialization pattern:

**Signal Handling and Graceful Shutdown**

The application implements proper signal handling for graceful shutdown:

  * [main.go\#L24-L38](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L24-L38): Context cancellation on interrupt signals
  * [main.go\#L104-L116](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L104-L116): Graceful HTTP and gRPC server shutdown with timeouts

Sources: [main.go\#L22-L116](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L22-L116) [main.go\#L40-L48](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L40-L48) [main.go\#L50-L66](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L50-L66) [main.go\#L68-L98](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/main.go#L68-L98)

## Docker Development

### Container Build Process

**Docker Commands**

| Command | Purpose |
| :--- | :--- |
| `make docker-build` | Build container image with tag `cryptovate/websocket-service:latest` |
| `make docker-run` | Run container with ports 8080 and 9090 exposed |

Sources: [Makefile\#L46-L53](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L46-L53)

## Development Best Practices

### File Organization

The codebase follows Go project layout conventions:

  * `main.go`: Application entry point and server initialization
  * `internal/`: Private application code not intended for import
  * `protos/`: Protocol Buffer service definitions
  * `gen/`: Generated code from Protocol Buffers
  * `config/`: Environment-specific configuration files

### Build Optimization

The build process uses optimization flags for production binaries:

  * [Makefile\#L12-L12](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L12-L12): `LDFLAGS := -ldflags "-s -w"` removes debug info and symbol tables
  * Binary name: `websocket-service` as defined in [Makefile\#L4-L4](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L4-L4)

Sources: [.gitignore\#L1-L46](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/.gitignore#L1-L46) [Makefile\#L4-L4](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L4-L4) [Makefile\#L12-L12](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L12-L12) [Makefile\#L27-L28](https://github.com/consentsam/websocket-integration-challange/blob/97cb3ae4/Makefile#L27-L28)
