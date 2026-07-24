package observability

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockHandlerCalled struct {
	called bool
}

func mockUnaryHandler(called *mockHandlerCalled) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		called.called = true
		return "response", nil
	}
}

type mockInvokerCalled struct {
	called bool
}

type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)       {}
func (m *mockServerStream) Context() context.Context     { return m.ctx }
func (m *mockServerStream) SendMsg(interface{}) error    { return nil }
func (m *mockServerStream) RecvMsg(interface{}) error    { return nil }

type mockClientStream struct {
	ctx context.Context
}

func (m *mockClientStream) Header() (metadata.MD, error) { return nil, nil }
func (m *mockClientStream) Trailer() metadata.MD         { return nil }
func (m *mockClientStream) CloseSend() error             { return nil }
func (m *mockClientStream) Context() context.Context     { return m.ctx }
func (m *mockClientStream) SendMsg(interface{}) error    { return nil }
func (m *mockClientStream) RecvMsg(interface{}) error    { return nil }

func mockUnaryInvoker(called *mockInvokerCalled) grpc.UnaryInvoker {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		called.called = true
		return nil
	}
}

func TestUnaryServerInterceptor_NilTracer(t *testing.T) {
	interceptor := UnaryServerInterceptor()
	require.NotNil(t, interceptor)

	handlerCalled := &mockHandlerCalled{}
	handler := mockUnaryHandler(handlerCalled)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	req := "request"
	resp, err := interceptor(context.Background(), req, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, handlerCalled.called)
}

func TestStreamServerInterceptor_NilTracer(t *testing.T) {
	interceptor := StreamServerInterceptor()
	require.NotNil(t, interceptor)

	assert.NotNil(t, interceptor)
}

func TestUnaryClientInterceptor_NilTracer(t *testing.T) {
	interceptor := UnaryClientInterceptor()
	require.NotNil(t, interceptor)

	invokerCalled := &mockInvokerCalled{}
	invoker := mockUnaryInvoker(invokerCalled)

	cc := &grpc.ClientConn{}
	req := "request"
	reply := new(string)

	err := interceptor(context.Background(), "/test.Service/Method", req, reply, cc, invoker)

	require.NoError(t, err)
	assert.True(t, invokerCalled.called)
}

func TestStreamClientInterceptor_NilTracer(t *testing.T) {
	interceptor := StreamClientInterceptor()
	require.NotNil(t, interceptor)

	assert.NotNil(t, interceptor)
}

func TestUnaryServerInterceptor_CallsHandler(t *testing.T) {
	interceptor := UnaryServerInterceptor()

	tests := []struct {
		name       string
		method     string
		req        interface{}
		handler    grpc.UnaryHandler
		expectResp interface{}
		expectErr  bool
	}{
		{
			name:       "calls handler with correct request",
			method:     "/bank.Account/GetBalance",
			req:        "account-123",
			handler:    mockUnaryHandler(&mockHandlerCalled{}),
			expectResp: "response",
			expectErr:  false,
		},
		{
			name:   "calls handler with different method",
			method: "/bank.Transfer/Create",
			req:    map[string]string{"from": "A", "to": "B"},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return "transfer-ok", nil
			},
			expectResp: "transfer-ok",
			expectErr:  false,
		},
		{
			name:   "handler returns nil response",
			method: "/test/Nil",
			req:    nil,
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
			expectResp: nil,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &grpc.UnaryServerInfo{FullMethod: tt.method}
			resp, err := interceptor(context.Background(), tt.req, info, tt.handler)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectResp, resp)
		})
	}
}

func TestUnaryServerInterceptor_HandlerError(t *testing.T) {
	interceptor := UnaryServerInterceptor()

	tests := []struct {
		name      string
		method    string
		handler   grpc.UnaryHandler
		expectErr bool
	}{
		{
			name:   "handler returns error",
			method: "/test/Error",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, errors.New("insufficient funds")
			},
			expectErr: true,
		},
		{
			name:   "handler returns nil error",
			method: "/test/Success",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return "ok", nil
			},
			expectErr: false,
		},
		{
			name:   "handler returns context error",
			method: "/test/ContextCancelled",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, context.Canceled
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &grpc.UnaryServerInfo{FullMethod: tt.method}
			_, err := interceptor(context.Background(), "req", info, tt.handler)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnaryClientInterceptor_CallsInvoker(t *testing.T) {
	interceptor := UnaryClientInterceptor()

	tests := []struct {
		name        string
		method      string
		req         interface{}
		invoker     grpc.UnaryInvoker
		expectReply interface{}
		expectErr   bool
	}{
		{
			name:   "calls invoker with correct method",
			method: "/bank.Account/GetBalance",
			req:    "account-123",
			invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				*reply.(*string) = "balance: 1000"
				return nil
			},
			expectReply: "balance: 1000",
			expectErr:   false,
		},
		{
			name:   "calls invoker with different method",
			method: "/bank.Transfer/Create",
			req:    map[string]string{"from": "A", "to": "B"},
			invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				*reply.(*string) = "transfer-id-456"
				return nil
			},
			expectReply: "transfer-id-456",
			expectErr:   false,
		},
		{
			name:   "invoker returns nil reply",
			method: "/test/NilReply",
			req:    nil,
			invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return nil
			},
			expectReply: "",
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &grpc.ClientConn{}
			reply := new(string)

			err := interceptor(context.Background(), tt.method, tt.req, reply, cc, tt.invoker)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectReply, *reply)
		})
	}
}

func TestUnaryClientInterceptor_InvokerError(t *testing.T) {
	interceptor := UnaryClientInterceptor()

	tests := []struct {
		name      string
		method    string
		invoker   grpc.UnaryInvoker
		expectErr bool
	}{
		{
			name:   "invoker returns error",
			method: "/test/Error",
			invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return errors.New("connection refused")
			},
			expectErr: true,
		},
		{
			name:   "invoker returns nil error",
			method: "/test/Success",
			invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return nil
			},
			expectErr: false,
		},
		{
			name:   "invoker returns deadline exceeded",
			method: "/test/Deadline",
			invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return context.DeadlineExceeded
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &grpc.ClientConn{}
			reply := new(string)

			err := interceptor(context.Background(), tt.method, nil, reply, cc, tt.invoker)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnaryServerInterceptor_ContextPropagation(t *testing.T) {
	interceptor := UnaryServerInterceptor()

	type ctxKey string
	key := ctxKey("test-key")

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		val := ctx.Value(key)
		return val, nil
	}

	ctx := context.WithValue(context.Background(), key, "test-value")
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Context"}
	resp, err := interceptor(ctx, "req", info, handler)

	require.NoError(t, err)
	assert.Equal(t, "test-value", resp)
}

func TestUnaryServerInterceptor_MetadataExtraction(t *testing.T) {
	interceptor := UnaryServerInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "no-metadata", nil
		}
		keys := make([]string, 0, len(md))
		for k := range md {
			keys = append(keys, k)
		}
		return keys, nil
	}

	md := metadata.New(map[string]string{"x-request-id": "123"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Metadata"}
	resp, err := interceptor(ctx, "req", info, handler)

	require.NoError(t, err)
	keys, ok := resp.([]string)
	require.True(t, ok)
	assert.Contains(t, keys, "x-request-id")
}

func TestUnaryServerInterceptor_WithTracer(t *testing.T) {
	shutdown, err := Init("test", "")
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	interceptor := UnaryServerInterceptor()
	require.NotNil(t, interceptor)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
	resp, err := interceptor(context.Background(), "req", info, handler)

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
	assert.True(t, handlerCalled)
}

func TestUnaryServerInterceptor_WithTracerAndMetadata(t *testing.T) {
	shutdown, err := Init("test", "")
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	interceptor := UnaryServerInterceptor()
	require.NotNil(t, interceptor)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	}

	md := metadata.New(map[string]string{"x-request-id": "test-123"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
	resp, err := interceptor(ctx, "req", info, handler)

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
	assert.True(t, handlerCalled)
}

func TestStreamServerInterceptor_WithTracer(t *testing.T) {
	shutdown, err := Init("test", "")
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	interceptor := StreamServerInterceptor()
	require.NotNil(t, interceptor)

	handlerCalled := false
	handler := func(srv interface{}, ss grpc.ServerStream) error {
		handlerCalled = true
		return nil
	}

	ss := &mockServerStream{ctx: context.Background()}
	info := &grpc.StreamServerInfo{FullMethod: "/test/StreamMethod"}
	err = interceptor(nil, ss, info, handler)

	require.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestUnaryClientInterceptor_WithTracer(t *testing.T) {
	shutdown, err := Init("test", "")
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	interceptor := UnaryClientInterceptor()
	require.NotNil(t, interceptor)

	invokerCalled := false
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		invokerCalled = true
		return nil
	}

	cc := &grpc.ClientConn{}
	err = interceptor(context.Background(), "/test/Method", "req", new(string), cc, invoker)

	require.NoError(t, err)
	assert.True(t, invokerCalled)
}

func TestStreamClientInterceptor_WithTracer(t *testing.T) {
	shutdown, err := Init("test", "")
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	interceptor := StreamClientInterceptor()
	require.NotNil(t, interceptor)

	streamerCalled := false
	streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		streamerCalled = true
		return &mockClientStream{ctx: ctx}, nil
	}

	cc := &grpc.ClientConn{}
	desc := &grpc.StreamDesc{StreamName: "test-stream"}
	stream, err := interceptor(context.Background(), desc, cc, "/test/StreamMethod", streamer)

	require.NoError(t, err)
	assert.NotNil(t, stream)
	assert.True(t, streamerCalled)
}
