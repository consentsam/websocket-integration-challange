# Build stage
FROM golang:1.22-alpine AS builder

# Install git and CA certificates (needed for private repos and HTTPS)
RUN apk --no-cache add git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./

# Create the expected directory structure for server-utils
RUN mkdir -p ../../pkg
COPY server-utils ../../pkg/server-utils

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o websocket-service .

# Final stage
FROM alpine:latest

# Install CA certificates, timezone data, and wget for health checks
RUN apk --no-cache add ca-certificates tzdata wget \
    && addgroup -g 1001 -S appgroup \
    && adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy timezone data and CA certificates
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from the builder stage
COPY --from=builder /app/websocket-service .

# Copy the configuration files
COPY --from=builder /app/config ./config

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Set environment variables
ENV ENVIRONMENT=production
ENV GIN_MODE=release

# Expose the HTTP and gRPC ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Set the entry point
ENTRYPOINT ["./websocket-service"]
