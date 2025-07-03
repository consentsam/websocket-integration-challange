# Makefile for the websocket service

# Variables
SERVICE_NAME := websocket-service
GO := go
PROTOC := protoc
DOCKER := docker
DOCKER_IMAGE := cryptovate/$(SERVICE_NAME)
DOCKER_TAG := latest

# Go build flags
LDFLAGS := -ldflags "-s -w"

# Directories
PROTO_DIR := protos
GEN_DIR := gen

# Main targets
.PHONY: all
all: clean deps proto build

.PHONY: run
run:
	$(GO) run main.go

.PHONY: build
build: proto
	$(GO) build $(LDFLAGS) -o $(SERVICE_NAME) main.go

.PHONY: clean
clean:
	rm -f $(SERVICE_NAME)
	rm -rf $(GEN_DIR)

.PHONY: deps
deps:
	$(GO) mod download
	$(GO) mod tidy

# Protocol Buffers
.PHONY: proto
proto:
	mkdir -p $(GEN_DIR)
	protoc --go_out=./gen ./protos/websocket/v1/api.proto  --go-grpc_out=./gen

# Docker
.PHONY: docker-build
docker-build:
	$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run:
	$(DOCKER) run -p 8080:8080 -p 9090:9090 $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: deploy-aws
deploy-aws:
	@echo "🚀 Deploying to AWS using CDK (proven approach)..."
	deploy/aws/deploy.sh

.PHONY: deploy-aws-destroy
deploy-aws-destroy:
	@echo "🗑️  Destroying AWS CDK stack..."
	cd deploy/aws/cdk && cdk destroy WebSocketServiceStack --force

.PHONY: deploy-aws-diff
deploy-aws-diff:
	@echo "📋 Showing CDK deployment diff..."
	cd deploy/aws/cdk && cdk diff WebSocketServiceStack

.PHONY: deploy-aws-logs
deploy-aws-logs:
	@echo "📋 Showing AWS service logs..."
	aws logs tail /aws/ecs/websocket-service --follow --region us-east-1

.PHONY: aws-update
aws-update:
	./aws-deploy.sh update

.PHONY: aws-destroy
aws-destroy:
	./aws-deploy.sh destroy

.PHONY: aws-status
aws-status:
	./aws-deploy.sh status

.PHONY: aws-logs
aws-logs:
	./aws-deploy.sh logs

.PHONY: aws-info
aws-info:
	./aws-deploy.sh info

.PHONY: aws-test
aws-test:
	./aws-deploy.sh test

# Testing
.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: test-coverage
test-coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Development
.PHONY: dev
dev:
	air -c .air.toml

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all            - Clean, download dependencies, generate proto files, and build"
	@echo "  run            - Run the service"
	@echo "  build          - Build the service"
	@echo "  clean          - Remove build artifacts"
	@echo "  deps           - Download dependencies"
	@echo "  proto          - Generate code from Protocol Buffers"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  dev            - Run with hot reload (requires air)"
	@echo ""
	@echo "Local deployment targets:"
	@echo "  deploy         - Deploy basic service with Docker Compose"
	@echo "  deploy-production - Deploy with Nginx reverse proxy"
	@echo "  deploy-test    - Test current deployment"
	@echo "  deploy-stop    - Stop all deployment services"
	@echo "  deploy-clean   - Clean up all containers and images"
	@echo "  deploy-logs    - Show deployment logs"
	@echo ""
	@echo "AWS CDK deployment targets (RECOMMENDED):"
	@echo "  deploy-aws     - Deploy to AWS using CDK (proven approach)"
	@echo "  deploy-aws-diff - Show what will change in AWS deployment"
	@echo "  deploy-aws-destroy - Destroy AWS CDK resources"
	@echo "  deploy-aws-logs - Show AWS service logs"
	@echo ""
	@echo "AWS legacy deployment targets:"
	@echo "  aws-deploy     - Deploy using legacy bash script"
	@echo "  aws-update     - Update AWS deployment with new code"
	@echo "  aws-destroy    - Destroy AWS deployment"
	@echo "  aws-status     - Show AWS service status"
	@echo "  aws-logs       - Show AWS service logs"
	@echo "  aws-info       - Show AWS deployment information"
	@echo "  aws-test       - Test AWS deployment"
	@echo ""
	@echo "  help           - Show this help"
