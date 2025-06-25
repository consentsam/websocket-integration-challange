package telemetry

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.22.0"

	"github.com/consentsam/websocket-integration-challange/internal/config"
)

// Init configures global telemetry providers and returns a mux exposing /metrics.
func Init(ctx context.Context, cfg *config.Config) (*http.ServeMux, error) {
	// Create Prometheus exporter for metrics
	prometheusExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	// Create meter provider with Prometheus exporter
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(prometheusExporter),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		)),
	)
	otel.SetMeterProvider(mp)

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		)),
	)
	otel.SetTracerProvider(tp)

	// Create HTTP mux with proper metrics endpoint
	mux := http.NewServeMux()
	mux.Handle(cfg.Metrics.Endpoint, promhttp.Handler())

	return mux, nil
}
