#!/bin/bash
set -e

# Quick fix for the current build error
echo "🔧 Quick fix for protobuf generation issue..."

# Ensure Go tools are in PATH
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
export PATH="$GOBIN:$PATH"

cd /workspace

# Install missing protoc Go plugins if needed
echo "📦 Installing protoc Go plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify plugins are available
if ! command -v protoc-gen-go &> /dev/null; then
    echo "❌ protoc-gen-go not found in PATH"
    exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "❌ protoc-gen-go-grpc not found in PATH"
    exit 1
fi

echo "✅ Protoc plugins available"

# Generate protobuf code
echo "🔄 Generating protobuf code..."
make proto

# Test build
echo "🔍 Testing build..."
make build

echo "✅ Fixed! You can now run 'make ci'" 