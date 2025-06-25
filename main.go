package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	websocketv1 "github.com/consentsam/websocket-integration-challange/gen/websocket/api/v1"
	"github.com/consentsam/websocket-integration-challange/internal/config"
	"github.com/consentsam/websocket-integration-challange/internal/handlers"
	"github.com/consentsam/websocket-integration-challange/internal/server"
	"github.com/consentsam/websocket-integration-challange/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// isTelemetryPhase5Enabled checks if telemetry phase 5 is enabled via environment variable.
// It follows standard boolean environment variable conventions:
// - Unset or empty: false (disabled)
// - Truthy values ("true", "1", "yes", "on", "enable"): true (enabled)
// - Falsy values ("false", "0", "no", "off", "disable"): false (disabled)
func isTelemetryPhase5Enabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("TELEMETRY_PHASE_5_ENABLED")))
	switch value {
	case "true", "1", "yes", "on", "enable":
		return true
	case "", "false", "0", "no", "off", "disable":
		return false
	default:
		// For any other values, default to false (disabled)
		return false
	}
}

func main() {

	// Create a context that is canceled when the program receives an interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
		// Give the server some time to gracefully shut down
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	// Load the configuration
	cfg, err := config.LoadConfig("websocket-service")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Bootstrap telemetry (phase 1)
	metricsMux, err := telemetry.Init(ctx, cfg)
	if err != nil {
		log.Fatalf("telemetry init failed: %v", err)
	}

	// Create the websocket handler
	websocketHandler := handlers.NewWebsocketHandler(ctx, cfg)
	defer websocketHandler.Close()

	// Create the gRPC server with optional telemetry interceptor (phase 5)
	grpcOpts := []grpc.ServerOption{}
	if isTelemetryPhase5Enabled() {
		grpcOpts = append(grpcOpts, grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
	}
	grpcServer := grpc.NewServer(grpcOpts...)
	websocketServer := server.NewServer(ctx, cfg, websocketHandler)
	websocketv1.RegisterWebsocketServiceServer(grpcServer, websocketServer)
	reflection.Register(grpcServer)

	// Start the gRPC server
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}
	go func() {
		log.Printf("Starting gRPC server on port %d", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Create the HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", websocketHandler.HandleWebsocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add metrics endpoint if enabled
	if cfg.Metrics.Enabled && metricsMux != nil {
		mux.Handle(cfg.Metrics.Endpoint, metricsMux)
	}

	// Start the HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: mux,
	}
	go func() {
		log.Printf("Starting HTTP server on port %d", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for the context to be canceled
	<-ctx.Done()

	// Shut down the HTTP server
	log.Println("Shutting down HTTP server...")
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpCancel()
	if err := httpServer.Shutdown(httpCtx); err != nil {
		log.Printf("Failed to shut down HTTP server: %v", err)
	}

	// Shut down the gRPC server
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()

	log.Println("Server shutdown complete")
}
