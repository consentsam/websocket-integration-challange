#!/bin/bash
set -euo pipefail

# =============================================================================
# CODEX Setup Script - WebSocket Service Bug Fixing
# Run this every time you start working on a new issue
# =============================================================================

echo "🚀 CODEX Environment Setup for WebSocket Service"

# Setup Go environment
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
export PATH="$GOBIN:$PATH"
mkdir -p "$GOPATH"/{bin,src,pkg}

# Navigate to project directory (CODEX clones to /workspace/websocket-integration-challange)
cd /workspace/websocket-integration-challange

# Install required Go tools (parallel for speed)
echo "📦 Installing Go development tools..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest &
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest &
go install honnef.co/go/tools/cmd/staticcheck@latest &
go install github.com/air-verse/air@latest &
go install golang.org/x/tools/cmd/goimports@latest &
wait

# Setup environment variables
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct
export WEBSOCKET_AUTH_SECRET="dev-secret-$(date +%s)"

# Download project dependencies
echo "📥 Installing project dependencies..."
go mod download
go mod tidy

# Generate protobuf code (fixes Bug #01)
echo "🔄 Generating protobuf code..."
make proto

# Verify everything works
echo "🔍 Testing build..."
make build

# Run full CI to ensure everything is ready
echo "✅ Running CI verification..."
make ci

echo ""
echo "============================================================================="
echo "🎉 CODEX ENVIRONMENT READY!"
echo "============================================================================="
echo "📁 Project: /workspace/websocket-integration-challange"
echo "🐹 Go: $(go version | cut -d' ' -f3)"
echo "🔧 Protoc: $(protoc --version 2>/dev/null || echo 'available')"
echo ""
echo "🚀 START BUG FIXING:"
echo "   1. Follow AGENTS.md workflow"
echo "   2. Create branch: git checkout -b bug/<id>-<title>"
echo "   3. Fix → Test → PR"
echo ""
echo "💡 QUICK COMMANDS:"
echo "   make dev        # Hot reload development"
echo "   make ci         # Run full CI pipeline"
echo "   make test-race  # Race condition testing"
echo "=============================================================================" 