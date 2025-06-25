#!/bin/bash
set -euo pipefail

# =============================================================================
# CODEX Development Environment Setup - WebSocket Service
# Assumes: Go 1.23.x preinstalled, Ubuntu 24.04
# =============================================================================

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; BLUE='\033[0;34m'; NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Tool versions
PROTOC_VERSION="28.3"

log_info "🚀 Setting up Go WebSocket service environment..."

# =============================================================================
# 1. VERIFY PREINSTALLED GO
# =============================================================================
if ! command -v go &> /dev/null; then
    log_error "Go not found! Please select Go 1.23.8 from CODEX preinstalled packages"
    exit 1
fi

GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
log_success "Using preinstalled Go ${GO_VERSION}"

# Setup Go environment
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
export PATH="$GOBIN:$PATH"
mkdir -p "$GOPATH"/{bin,src,pkg}

# =============================================================================
# 2. INSTALL PROTOC & GO TOOLS
# =============================================================================
log_info "🔧 Installing protoc..."
if ! command -v protoc &> /dev/null; then
    cd /tmp
    wget -q "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip"
    unzip -q "protoc-${PROTOC_VERSION}-linux-x86_64.zip"
    sudo cp bin/protoc /usr/local/bin/ && sudo cp -r include/* /usr/local/include/
    rm -rf bin include readme.txt "protoc-${PROTOC_VERSION}-linux-x86_64.zip"
fi

log_info "🛠️ Installing Go development tools..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest &
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest &
go install honnef.co/go/tools/cmd/staticcheck@latest &
go install github.com/cosmtrek/air@latest &
go install golang.org/x/tools/cmd/goimports@latest &
wait

# =============================================================================
# 3. SETUP ENVIRONMENT & PROJECT
# =============================================================================
export GO111MODULE=on GOPROXY=https://proxy.golang.org,direct
export WEBSOCKET_AUTH_SECRET="dev-secret-$(date +%s)"

cd /workspace

log_info "📦 Setting up project..."
[ -f "go.mod" ] && { go mod download; go mod tidy; }
[ -f "Makefile" ] && make proto

log_info "🔍 Testing build..."
if [ -f "Makefile" ] && ! make build &> /dev/null; then
    log_error "Build failed! Check your setup."
    exit 1
fi

# =============================================================================
# 4. VERIFICATION & SUMMARY
# =============================================================================
log_success "🎉 Setup completed!"
echo "============================================================================="
echo "📋 ENVIRONMENT:"
echo "   • Go: $(go version | cut -d' ' -f3)"
echo "   • Protoc: $(protoc --version 2>/dev/null || echo 'installed')"
echo "   • Project: /workspace/"
echo
echo "🚀 READY TO CODE:"
echo "   make dev        # Start hot reload"
echo "   make ci         # Run full CI"
echo "   # Follow AGENTS.md for bug-fixing workflow"
echo "=============================================================================" 