#!/bin/bash

# AWS CDK Deployment Script for WebSocket Service
# This script follows the proven deployment approach from the previous project

set -e

# Configuration
PROJECT_NAME="websocket-service"
AWS_REGION="${AWS_DEFAULT_REGION:-us-east-1}"
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text 2>/dev/null || echo "")

if [ -z "$AWS_ACCOUNT_ID" ]; then
    echo "❌ Error: AWS credentials not configured properly"
    echo "Please run: aws configure"
    exit 1
fi

echo "🚀 Starting CDK deployment for $PROJECT_NAME"
echo "📍 Region: $AWS_REGION"
echo "🔐 AWS Account ID: $AWS_ACCOUNT_ID"

# Step 1: Install AWS CDK if not present
if ! command -v cdk &> /dev/null; then
    echo "📦 Installing AWS CDK..."
    npm install -g aws-cdk
fi

# Step 2: Build the application
echo "🔨 Building application..."

# Save current directory and determine project root
DEPLOY_DIR=$(pwd)

# Find the project root by looking for go.mod file
if [[ "$DEPLOY_DIR" == */deploy/aws ]]; then
    PROJECT_ROOT="${DEPLOY_DIR%/deploy/aws}"
else
    # Fallback: find go.mod in parent directories
    PROJECT_ROOT="$DEPLOY_DIR"
    while [[ "$PROJECT_ROOT" != "/" ]] && [[ ! -f "$PROJECT_ROOT/go.mod" ]]; do
        PROJECT_ROOT="$(dirname "$PROJECT_ROOT")"
    done
fi

echo "📁 Project root: $PROJECT_ROOT"
cd "$PROJECT_ROOT"

echo "📍 Building from: $(pwd)"

# Generate protobuf files if needed
mkdir -p gen
if [ -f "protos/websocket/v1/api.proto" ]; then
    echo "📋 Generating protobuf files..."
    protoc --go_out=./gen ./protos/websocket/v1/api.proto --go-grpc_out=./gen
else
    echo "⚠️ Proto file not found, using existing generated files"
fi

# Download dependencies
echo "📥 Downloading Go dependencies..."
go mod download

# Build the Go application
echo "🔨 Building Go binary..."
go build -ldflags "-s -w" -o websocket-service main.go
echo "✅ Application built successfully"

# Step 3: Build and push Docker image (from project root)
echo "🐳 Building and pushing Docker image..."

# Get ECR repository URI (create if doesn't exist)
ECR_REPOSITORY_NAME="websocket-service"
ECR_URI="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY_NAME"

# Create ECR repository if it doesn't exist
aws ecr describe-repositories --repository-names $ECR_REPOSITORY_NAME --region $AWS_REGION > /dev/null 2>&1 || {
    echo "📦 Creating ECR repository..."
    aws ecr create-repository \
        --repository-name $ECR_REPOSITORY_NAME \
        --region $AWS_REGION \
        --image-scanning-configuration scanOnPush=true
}

# Login to ECR
echo "🔑 Logging into ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $ECR_URI

# Build Docker image from project root
echo "🔨 Building Docker image..."
docker build -t $PROJECT_NAME "$PROJECT_ROOT"
docker tag $PROJECT_NAME:latest $ECR_URI:latest

echo "📤 Pushing Docker image to ECR..."
docker push $ECR_URI:latest

echo "✅ Docker image pushed to ECR: $ECR_URI:latest"

# Return to deploy directory for CDK
cd "$DEPLOY_DIR"

# Step 4: Deploy infrastructure using CDK
echo "🏗️ Deploying infrastructure with CDK..."

# Go to the deploy/aws directory (where cdk.json is located)
AWS_DEPLOY_DIR="$PROJECT_ROOT/deploy/aws"
cd "$AWS_DEPLOY_DIR"

# Download CDK dependencies
echo "📥 Installing CDK dependencies..."
cd cdk && go mod download && cd ..

# Bootstrap CDK if needed (from the directory containing cdk.json)
cdk bootstrap aws://$AWS_ACCOUNT_ID/$AWS_REGION || echo "CDK already bootstrapped"

# Deploy the stack
echo "🚀 Deploying CloudFormation stack..."
cdk deploy WebSocketServiceStack --require-approval never --outputs-file outputs.json

# Step 5: Display deployment outputs
echo ""
echo "🎉 Deployment completed successfully!"
echo ""

if [ -f outputs.json ]; then
    echo "📋 Deployment Outputs:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    SERVICE_URL=$(cat outputs.json | jq -r '.WebSocketServiceStack.ServiceURL // empty')
    WEBSOCKET_URL=$(cat outputs.json | jq -r '.WebSocketServiceStack.WebSocketURL // empty')
    HEALTH_URL=$(cat outputs.json | jq -r '.WebSocketServiceStack.HealthCheckURL // empty')
    METRICS_URL=$(cat outputs.json | jq -r '.WebSocketServiceStack.MetricsURL // empty')
    
    if [ ! -z "$SERVICE_URL" ]; then
        echo "🌐 Service URL: $SERVICE_URL"
    fi
    if [ ! -z "$WEBSOCKET_URL" ]; then
        echo "🔌 WebSocket URL: $WEBSOCKET_URL (NO AUTHENTICATION REQUIRED)"
    fi
    if [ ! -z "$HEALTH_URL" ]; then
        echo "📊 Health Check: $HEALTH_URL"
    fi
    if [ ! -z "$METRICS_URL" ]; then
        echo "📈 Metrics: $METRICS_URL"
    fi
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    if [ ! -z "$WEBSOCKET_URL" ]; then
        echo "📝 WebSocket Connection Examples:"
        echo "   JavaScript: const ws = new WebSocket('$WEBSOCKET_URL');"
        echo "   Python: websocket.create_connection('$WEBSOCKET_URL')"
        echo "   cURL: curl -i -N -H 'Connection: Upgrade' -H 'Upgrade: websocket' '$WEBSOCKET_URL'"
        echo ""
    fi
    
    echo "✅ WebSocket service is deployed with ZERO AUTHENTICATION"
    echo "🔒 Security: CORS enabled (*), Rate limiting disabled"
    echo "📊 Monitoring: CloudWatch logs and container insights enabled"
    echo ""
    echo "⏳ Note: It may take 2-3 minutes for the service to be fully healthy."
else
    echo "⚠️ CDK outputs file not found. Check CloudFormation console for outputs."
fi

echo "🔍 Monitor deployment:"
echo "   AWS Console: https://console.aws.amazon.com/ecs/home?region=$AWS_REGION#/clusters"
echo "   Logs: aws logs tail /aws/ecs/websocket-service --follow --region $AWS_REGION" 