package telemetry

import (
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Meter is the global meter used by the service.
var Meter = otel.Meter("websocket-service")

// Counter creates an int64 counter with the provided name and options.
func Counter(name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return Meter.Int64Counter(name, opts...)
}

// Histogram creates an int64 histogram with the provided name and options.
func Histogram(name string, opts ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	return Meter.Int64Histogram(name, opts...)
}

// MustCounter creates an int64 counter and panics on error. Only use during initialization.
func MustCounter(name string, opts ...metric.Int64CounterOption) metric.Int64Counter {
	counter, err := Counter(name, opts...)
	if err != nil {
		log.Fatalf("Failed to create counter %s: %v", name, err)
	}
	return counter
}

// MustHistogram creates an int64 histogram and panics on error. Only use during initialization.
func MustHistogram(name string, opts ...metric.Int64HistogramOption) metric.Int64Histogram {
	histogram, err := Histogram(name, opts...)
	if err != nil {
		log.Fatalf("Failed to create histogram %s: %v", name, err)
	}
	return histogram
}
