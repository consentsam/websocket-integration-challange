package telemetry

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.22.0"

	"github.com/consentsam/websocket-integration-challange/internal/config"
)

// Init configures global telemetry providers and returns a mux exposing /metrics.
func Init(ctx context.Context, cfg *config.Config) (*http.ServeMux, error) {
	if strings.ToLower(os.Getenv("TELEMETRY_PHASE_1_ENABLED")) == "false" {
		return nil, nil
	}

	log.Printf("telemetry phase 1 enabled")

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		)),
	)
	otel.SetTracerProvider(tp)

	mp := sdkmetric.NewMeterProvider()
	otel.SetMeterProvider(mp)

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.Metrics.Endpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return mux, nil
}
