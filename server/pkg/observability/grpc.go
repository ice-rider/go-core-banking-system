package observability

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if globalTracer != nil {
			ctx, span := globalTracer.Start(ctx, info.FullMethod)
			defer span.End()
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				_ = md
			}
		}
		return handler(ctx, req)
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if globalTracer != nil {
			_, span := globalTracer.Start(ss.Context(), info.FullMethod)
			defer span.End()
		}
		return handler(srv, ss)
	}
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if globalTracer != nil {
			_, span := globalTracer.Start(ctx, method)
			defer span.End()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if globalTracer != nil {
			_, span := globalTracer.Start(ctx, method)
			defer span.End()
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func GRPCServerInterceptors() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	return UnaryServerInterceptor(), StreamServerInterceptor()
}

func GRPCClientInterceptors() (grpc.UnaryClientInterceptor, grpc.StreamClientInterceptor) {
	return UnaryClientInterceptor(), StreamClientInterceptor()
}

func ServerStatsHandler() stats.Handler {
	return nil
}

func ClientStatsHandler() stats.Handler {
	return nil
}
