package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Meter is the global meter used by the service.
var Meter = otel.Meter("websocket-service")

// Counter creates an int64 counter with the provided name and options.
func Counter(name string, opts ...metric.InstrumentOption) (metric.Int64Counter, error) {
	return Meter.Int64Counter(name, opts...)
}

// Histogram creates an int64 histogram with the provided name and options.
func Histogram(name string, opts ...metric.InstrumentOption) (metric.Int64Histogram, error) {
	return Meter.Int64Histogram(name, opts...)
}
