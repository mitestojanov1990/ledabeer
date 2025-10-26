package logging

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(logger *Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		ctx = WithRequestID(ctx)
		ctx = WithLogger(ctx, logger)

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		l := FromContext(ctx)
		if err != nil {
			l.Error("gRPC request failed",
				String("grpc.method", info.FullMethod),
				Int64("grpc.duration_ms", duration.Milliseconds()),
				String("error", err.Error()),
			)
		} else {
			l.Info("gRPC request completed",
				String("grpc.method", info.FullMethod),
				Int64("grpc.duration_ms", duration.Milliseconds()),
			)
		}

		return resp, err
	}
}

func WebSocketHandler(logger *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("websocket_connection",
			String("remote_addr", r.RemoteAddr),
		)
		next.ServeHTTP(w, r)
	})
}
