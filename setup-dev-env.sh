#!/bin/bash
set -euo pipefail

# =============================================================================
# Development Environment Setup Script for WebSocket Service
# Target: CODEX Environment (Ubuntu 24.04, /workspace/)
# =============================================================================

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Tool versions
GO_VERSION="1.22.8"
PROTOC_VERSION="28.3"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_command() {
    if command -v "$1" &> /dev/null; then
        log_success "$1 is available"
        return 0
    else
        log_warning "$1 is not available"
        return 1
    fi
}

# =============================================================================
# 1. SYSTEM DEPENDENCIES
# =============================================================================
log_info "🚀 Starting development environment setup..."

# Update package index
log_info "📦 Updating package index..."
sudo apt-get update -qq

# Install basic dependencies
log_info "📦 Installing system dependencies..."
sudo apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    make \
    unzip \
    ca-certificates

# =============================================================================
# 2. GO INSTALLATION
# =============================================================================
log_info "🐹 Installing Go ${GO_VERSION}..."

# Check if Go is already installed with correct version
if command -v go &> /dev/null; then
    CURRENT_GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+\.[0-9]+')
    if [[ "$CURRENT_GO_VERSION" == "$GO_VERSION" ]]; then
        log_success "Go ${GO_VERSION} is already installed"
    else
        log_warning "Go ${CURRENT_GO_VERSION} found, upgrading to ${GO_VERSION}"
        sudo rm -rf /usr/local/go
    fi
fi

# Install Go if not present or version mismatch
if ! command -v go &> /dev/null || [[ "$(go version | grep -oP 'go\K[0-9]+\.[0-9]+\.[0-9]+')" != "$GO_VERSION" ]]; then
    cd /tmp
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"
fi

# Setup Go environment
export PATH="/usr/local/go/bin:$PATH"
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
export PATH="$GOBIN:$PATH"

# Add to bashrc for persistence
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
    echo 'export GOPATH="$HOME/go"' >> ~/.bashrc
    echo 'export GOBIN="$GOPATH/bin"' >> ~/.bashrc
    echo 'export PATH="$GOBIN:$PATH"' >> ~/.bashrc
fi

# Create GOPATH directories
mkdir -p "$GOPATH"/{bin,src,pkg}

log_success "Go ${GO_VERSION} installed successfully"

# =============================================================================
# 3. PROTOCOL BUFFERS INSTALLATION
# =============================================================================
log_info "🔧 Installing Protocol Buffers compiler..."

# Install protoc
if ! command -v protoc &> /dev/null; then
    cd /tmp
    wget -q "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip"
    unzip -q "protoc-${PROTOC_VERSION}-linux-x86_64.zip"
    sudo cp bin/protoc /usr/local/bin/
    sudo cp -r include/* /usr/local/include/
    rm -rf bin include readme.txt "protoc-${PROTOC_VERSION}-linux-x86_64.zip"
    log_success "protoc ${PROTOC_VERSION} installed"
else
    log_success "protoc is already installed"
fi

# Install Go protobuf plugins
log_info "🔌 Installing Go protobuf plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# =============================================================================
# 4. DEVELOPMENT TOOLS
# =============================================================================
log_info "🛠️ Installing development tools..."

# Install staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install air for hot reload
go install github.com/cosmtrek/air@latest

# Install other useful tools
go install golang.org/x/tools/cmd/goimports@latest

# =============================================================================
# 5. ENVIRONMENT VARIABLES SETUP
# =============================================================================
log_info "⚙️ Setting up environment variables..."

# Create .env template for development
cat > /workspace/.env.template << 'EOF'
# Development Environment Variables
# Copy this file to .env and fill in the values

# WebSocket Authentication Secret (required for development/production)
WEBSOCKET_AUTH_SECRET=your-development-secret-here

# Go Environment
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

# Development Settings
export LOG_LEVEL=debug
export ENVIRONMENT=local
EOF

# Set default development environment variables
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org
export WEBSOCKET_AUTH_SECRET="dev-secret-$(date +%s)"

log_success "Environment variables configured"

# =============================================================================
# 6. PROJECT INITIALIZATION
# =============================================================================
log_info "📁 Initializing project in /workspace/..."

# Navigate to workspace
cd /workspace

# Download Go dependencies
if [ -f "go.mod" ]; then
    log_info "📥 Downloading Go dependencies..."
    go mod download
    go mod tidy
    log_success "Go dependencies downloaded"
else
    log_warning "go.mod not found - skipping dependency download"
fi

# Generate protobuf code
if [ -f "Makefile" ] && make -n proto &> /dev/null; then
    log_info "🔄 Generating protobuf code..."
    make proto
    log_success "Protobuf code generated"
else
    log_warning "Makefile or proto target not found - skipping protobuf generation"
fi

# =============================================================================
# 7. BUILD VERIFICATION
# =============================================================================
log_info "🔍 Verifying project build..."

if [ -f "Makefile" ]; then
    # Test build
    if make build &> /dev/null; then
        log_success "Project builds successfully"
    else
        log_error "Project build failed"
        exit 1
    fi
    
    # Test dependencies
    if make deps &> /dev/null; then
        log_success "Dependencies are satisfied"
    else
        log_warning "Some dependencies may be missing"
    fi
else
    log_warning "Makefile not found - skipping build verification"
fi

# =============================================================================
# 8. VERIFICATION CHECKLIST
# =============================================================================
log_info "✅ Running verification checklist..."

# Check all required tools
TOOLS=("go" "protoc" "staticcheck" "air" "goimports" "make" "git")
ALL_TOOLS_OK=true

for tool in "${TOOLS[@]}"; do
    if ! check_command "$tool"; then
        ALL_TOOLS_OK=false
    fi
done

# Check Go version
GO_ACTUAL_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+\.[0-9]+')
if [[ "$GO_ACTUAL_VERSION" == "$GO_VERSION" ]]; then
    log_success "Go version: $GO_ACTUAL_VERSION ✓"
else
    log_warning "Go version mismatch: expected $GO_VERSION, got $GO_ACTUAL_VERSION"
    ALL_TOOLS_OK=false
fi

# Check protoc plugins
if command -v protoc-gen-go &> /dev/null && command -v protoc-gen-go-grpc &> /dev/null; then
    log_success "Protobuf Go plugins installed ✓"
else
    log_error "Protobuf Go plugins missing"
    ALL_TOOLS_OK=false
fi

# Check environment
if [[ -n "${GOPATH:-}" ]] && [[ -n "${GOBIN:-}" ]]; then
    log_success "Go environment configured ✓"
else
    log_error "Go environment not properly configured"
    ALL_TOOLS_OK=false
fi

# =============================================================================
# 9. SETUP SUMMARY
# =============================================================================
echo
echo "============================================================================="
if $ALL_TOOLS_OK; then
    log_success "🎉 Development environment setup completed successfully!"
else
    log_error "❌ Setup completed with some issues. Please review the warnings above."
fi
echo "============================================================================="
echo
echo "📋 QUICK REFERENCE:"
echo "   • Project directory: /workspace/"
echo "   • Go version: $(go version | cut -d' ' -f3)"
echo "   • Protoc version: $(protoc --version)"
echo "   • GOPATH: $GOPATH"
echo "   • GOBIN: $GOBIN"
echo
echo "🚀 NEXT STEPS:"
echo "   1. Set up authentication secret:"
echo "      export WEBSOCKET_AUTH_SECRET=\"your-secret-here\""
echo "   2. Start development:"
echo "      make dev                    # Hot reload development"
echo "      make ci                     # Run full CI pipeline"
echo "   3. Follow bug-fixing workflow in AGENTS.md"
echo
echo "📖 For detailed workflow instructions, see:"
echo "   • AGENTS.md - Bug fixing procedures"
echo "   • docs/development/branch-conventions.md - Branching guide"
echo "============================================================================="

# Source bashrc to make environment available in current session
log_info "🔄 Reloading shell environment..."
source ~/.bashrc 2>/dev/null || true

exit 0 