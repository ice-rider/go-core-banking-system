package resilience

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryClientInterceptor returns a gRPC unary client interceptor with timeout + circuit breaker.
func UnaryClientInterceptor(cb *CircuitBreaker, timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		_, err := cb.Execute(ctx, timeout, func(callCtx context.Context) (any, error) {
			err := invoker(callCtx, method, req, reply, cc, opts...)
			return nil, err
		})
		return err
	}
}

// StreamClientInterceptor returns a gRPC stream client interceptor with circuit breaker.
func StreamClientInterceptor(cb *CircuitBreaker) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		result, err := cb.Execute(ctx, 30*time.Second, func(callCtx context.Context) (any, error) {
			return streamer(callCtx, desc, cc, method, opts...)
		})
		if err != nil {
			return nil, err
		}
		return result.(grpc.ClientStream), nil
	}
}

// IsRetryable determines if an error is retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	switch code := status.Code(err); code {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}
