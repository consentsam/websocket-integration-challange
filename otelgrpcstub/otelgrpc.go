package otelgrpc

import (
	"context"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor provides a minimal gRPC interceptor that starts and
// ends a span around each unary call. This is a lightweight stub used in the
// CODEX environment where the full otelgrpc module is unavailable.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tracer := otel.Tracer("grpc-server")
		ctx, span := tracer.Start(ctx, info.FullMethod)
		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
		}
		span.End()
		return resp, err
	}
}
