package telemetry

import (
	"log"
	"net/http"

	"go.opentelemetry.io/otel/metric"
)

var panicTotal metric.Int64Counter

func init() {
	var err error
	panicTotal, err = Counter("go_panic_total")
	if err != nil {
		log.Printf("telemetry counter init failed: %v", err)
	}
}

// RecoveryMiddleware recovers from panics in HTTP handlers and records a metric.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if panicTotal != nil {
					panicTotal.Add(r.Context(), 1)
				}
				log.Printf("panic recovered: %v", rec)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
