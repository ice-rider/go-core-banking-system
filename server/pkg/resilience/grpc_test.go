package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryClientInterceptor_Success(t *testing.T) {
	cb := New(DefaultConfig("test"))
	interceptor := UnaryClientInterceptor(cb, time.Second)

	called := false
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		called = true
		return nil
	}

	err := interceptor(context.Background(), "/test/Method", nil, nil, nil, invoker)

	require.NoError(t, err)
	assert.True(t, called)
}

func TestUnaryClientInterceptor_Error(t *testing.T) {
	cb := New(DefaultConfig("test"))
	interceptor := UnaryClientInterceptor(cb, time.Second)

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return errors.New("rpc error")
	}

	err := interceptor(context.Background(), "/test/Method", nil, nil, nil, invoker)

	require.Error(t, err)
	assert.EqualError(t, err, "rpc error")
}

func TestUnaryClientInterceptor_Timeout(t *testing.T) {
	cb := New(DefaultConfig("test"))
	interceptor := UnaryClientInterceptor(cb, 50*time.Millisecond)

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return nil
		}
	}

	err := interceptor(context.Background(), "/test/Method", nil, nil, nil, invoker)

	require.Error(t, err)
}

func TestStreamClientInterceptor_Success(t *testing.T) {
	cb := New(DefaultConfig("test"))
	interceptor := StreamClientInterceptor(cb)

	dummyStream := &mockClientStream{}
	streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return dummyStream, nil
	}

	stream, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/test/Method", streamer)

	assert.NoError(t, err)
	assert.Equal(t, dummyStream, stream)
}

type mockClientStream struct {
	grpc.ClientStream
}

func (m *mockClientStream) SendMsg(msg any) error   { return nil }
func (m *mockClientStream) RecvMsg(msg any) error   { return nil }
func (m *mockClientStream) CloseSend() error         { return nil }

func TestStreamClientInterceptor_Error(t *testing.T) {
	cb := New(DefaultConfig("test"))
	interceptor := StreamClientInterceptor(cb)

	streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, errors.New("stream error")
	}

	_, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/test/Method", streamer)

	require.Error(t, err)
	assert.EqualError(t, err, "stream error")
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"unavailable", status.Error(codes.Unavailable, "unavailable"), true},
		{"deadline exceeded", status.Error(codes.DeadlineExceeded, "timeout"), true},
		{"resource exhausted", status.Error(codes.ResourceExhausted, "exhausted"), true},
		{"not found", status.Error(codes.NotFound, "not found"), false},
		{"invalid argument", status.Error(codes.InvalidArgument, "invalid"), false},
		{"generic error", errors.New("generic"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRetryable(tt.err))
		})
	}
}
