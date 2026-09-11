package obs

import (
	"context"

	"go.uber.org/zap"
)

type requestIDKey struct{}

// WithRequestID returns ctx carrying the request's correlation id.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestID returns the id set by the request id middleware, or "".
func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}

// Logger returns the process logger tagged with the request id from ctx.
func Logger(ctx context.Context) *zap.Logger {
	requestID := RequestID(ctx)
	if requestID == "" {
		return zap.L()
	}
	return zap.L().With(zap.String("request_id", requestID))
}
