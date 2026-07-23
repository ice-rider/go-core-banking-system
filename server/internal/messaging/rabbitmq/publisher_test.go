package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPublisherChannel struct {
	publishFunc func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	closeFunc   func() error
}

func (m *mockPublisherChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	if m.publishFunc != nil {
		return m.publishFunc(exchange, key, mandatory, immediate, msg)
	}
	return nil
}

func (m *mockPublisherChannel) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

type mockPublisherConnection struct {
	channelFunc func() (*amqp.Channel, error)
	closeFunc   func() error
	isClosed    bool
}

func (m *mockPublisherConnection) Channel() (*amqp.Channel, error) {
	if m.channelFunc != nil {
		return m.channelFunc()
	}
	return nil, nil
}

func (m *mockPublisherConnection) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	m.isClosed = true
	return nil
}

func (m *mockPublisherConnection) IsClosed() bool {
	return m.isClosed
}

func setupPublisherMockDialer(dialFunc func(string) (AMQPConnection, error)) func() {
	orig := NewDialer
	NewDialer = func() Dialer { return &mockDialer{dialFunc: dialFunc} }
	return func() { NewDialer = orig }
}

func createTestConnection(t *testing.T) *Connection {
	t.Helper()
	conn := &mockPublisherConnection{}
	restore := setupPublisherMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	t.Cleanup(restore)
	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)
	return c
}

func TestNewPublisher_Success(t *testing.T) {
	c := createTestConnection(t)

	p := NewPublisher(c, "test-exchange")

	require.NotNil(t, p)
	assert.Equal(t, "test-exchange", p.exchange)
	assert.False(t, p.closed)
	assert.Nil(t, p.ch)
	assert.Equal(t, c, p.conn)
}

func TestNewPublisher_DefaultChannelError(t *testing.T) {
	conn := &mockConnection{
		channelFunc: func() (*amqp.Channel, error) {
			return nil, errors.New("channel error")
		},
	}
	restore := setupPublisherMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	p := NewPublisher(c, "test-exchange")

	err = p.Publish(context.Background(), "routing.key", []byte("data"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open channel")
}

func TestPublish_Success(t *testing.T) {
	mockCh := &mockPublisherChannel{}

	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return mockCh, nil
		},
	}

	body := []byte(`{"key":"value"}`)
	var publishedMsg amqp.Publishing

	mockCh.publishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		publishedMsg = msg
		return nil
	}

	err := p.Publish(context.Background(), "routing.key", body)

	require.NoError(t, err)
	assert.Equal(t, body, publishedMsg.Body)
	assert.Equal(t, "application/json", publishedMsg.ContentType)
}

func TestPublish_Success_VerifiesArguments(t *testing.T) {
	tests := []struct {
		name       string
		exchange   string
		routingKey string
		body       []byte
	}{
		{
			name:       "simple message",
			exchange:   "events",
			routingKey: "order.created",
			body:       []byte(`{"event":"created"}`),
		},
		{
			name:       "nested routing key",
			exchange:   "logs",
			routingKey: "app.service.method",
			body:       []byte("log message"),
		},
		{
			name:       "unicode body",
			exchange:   "notifications",
			routingKey: "user.notify",
			body:       []byte(`{"msg":"Привет мир"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCh := &mockPublisherChannel{}
			c := createTestConnection(t)
			p := &Publisher{
				conn:     c,
				exchange: tt.exchange,
				newChannel: func() (PublisherChannel, error) {
					return mockCh, nil
				},
			}

			var capturedExchange, capturedKey string
			var capturedBody []byte
			mockCh.publishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
				capturedExchange = exchange
				capturedKey = key
				capturedBody = msg.Body
				return nil
			}

			err := p.Publish(context.Background(), tt.routingKey, tt.body)

			require.NoError(t, err)
			assert.Equal(t, tt.exchange, capturedExchange)
			assert.Equal(t, tt.routingKey, capturedKey)
			assert.Equal(t, tt.body, capturedBody)
		})
	}
}

func TestPublish_CreatesChannelOnce(t *testing.T) {
	var channelCallCount int
	mockCh := &mockPublisherChannel{
		publishFunc: func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
			return nil
		},
	}

	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			channelCallCount++
			return mockCh, nil
		},
	}

	err := p.Publish(context.Background(), "key1", []byte("data1"))
	require.NoError(t, err)

	err = p.Publish(context.Background(), "key2", []byte("data2"))
	require.NoError(t, err)

	assert.Equal(t, 1, channelCallCount)
}

func TestPublish_ConnectionClosed(t *testing.T) {
	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return nil, errors.New("connection is closed")
		},
	}

	err := p.Publish(context.Background(), "routing.key", []byte("data"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection is closed")
}

func TestPublish_ChannelError(t *testing.T) {
	mockCh := &mockPublisherChannel{
		publishFunc: func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
			return errors.New("channel is in use")
		},
	}

	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return mockCh, nil
		},
	}

	err := p.Publish(context.Background(), "routing.key", []byte("data"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "channel is in use")
}

func TestPublish_InvalidRoutingKey(t *testing.T) {
	mockCh := &mockPublisherChannel{
		publishFunc: func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
			return &amqp.Error{
				Code:   404,
				Reason: "NO_ROUTE",
			}
		},
	}

	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return mockCh, nil
		},
	}

	err := p.Publish(context.Background(), "invalid..routing..key", []byte("data"))

	require.Error(t, err)
	var amqpErr *amqp.Error
	assert.True(t, errors.As(err, &amqpErr))
	assert.Equal(t, 404, amqpErr.Code)
}

func TestPublish_EmptyBody(t *testing.T) {
	mockCh := &mockPublisherChannel{}
	var publishedMsg amqp.Publishing
	mockCh.publishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		publishedMsg = msg
		return nil
	}

	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return mockCh, nil
		},
	}

	err := p.Publish(context.Background(), "routing.key", []byte{})

	require.NoError(t, err)
	assert.Equal(t, []byte{}, publishedMsg.Body)
}

func TestPublish_NilBody(t *testing.T) {
	mockCh := &mockPublisherChannel{}
	var publishedMsg amqp.Publishing
	mockCh.publishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		publishedMsg = msg
		return nil
	}

	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return mockCh, nil
		},
	}

	err := p.Publish(context.Background(), "routing.key", nil)

	require.NoError(t, err)
	assert.Nil(t, publishedMsg.Body)
}

func TestPublish_ClosedPublisher(t *testing.T) {
	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		closed:   true,
	}

	err := p.Publish(context.Background(), "routing.key", []byte("data"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "publisher is closed")
}

func TestPublish_ContextCanceled(t *testing.T) {
	mockCh := &mockPublisherChannel{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return mockCh, nil
		},
	}

	mockCh.publishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		return ctx.Err()
	}

	err := p.Publish(ctx, "routing.key", []byte("data"))

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestPublisherClose_Success(t *testing.T) {
	mockCh := &mockPublisherChannel{}
	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return mockCh, nil
		},
	}

	err := p.Publish(context.Background(), "test-key", []byte("data"))
	require.NoError(t, err)

	err = p.Close()

	require.NoError(t, err)
	assert.True(t, p.closed)
}

func TestPublisherClose_AlreadyClosed(t *testing.T) {
	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		closed:   true,
	}

	err := p.Close()

	require.NoError(t, err)
	assert.True(t, p.closed)
}

func TestPublisherClose_ChannelCloseError(t *testing.T) {
	mockCh := &mockPublisherChannel{
		closeFunc: func() error {
			return errors.New("failed to close channel")
		},
	}

	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return mockCh, nil
		},
	}

	err := p.Publish(context.Background(), "test-key", []byte("data"))
	require.NoError(t, err)

	err = p.Close()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to close channel")
}

func TestPublisherClose_NilChannel(t *testing.T) {
	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
	}

	err := p.Close()

	require.NoError(t, err)
	assert.True(t, p.closed)
}

func TestPublish_ChannelConnectionFailure(t *testing.T) {
	c := createTestConnection(t)
	p := &Publisher{
		conn:     c,
		exchange: "test-exchange",
		newChannel: func() (PublisherChannel, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	err := p.Publish(context.Background(), "routing.key", []byte("data"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open channel")
}

func TestPublish_BodyPreserved(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{"json object", []byte(`{"event":"transfer","amount":100}`)},
		{"json array", []byte(`[1,2,3]`)},
		{"empty string", []byte("")},
		{"binary data", []byte{0x00, 0x01, 0x02, 0xFF}},
		{"large payload", make([]byte, 1024*1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCh := &mockPublisherChannel{}
			var capturedBody []byte
			mockCh.publishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
				capturedBody = msg.Body
				return nil
			}

			c := createTestConnection(t)
			p := &Publisher{
				conn:     c,
				exchange: "test-exchange",
				newChannel: func() (PublisherChannel, error) {
					return mockCh, nil
				},
			}

			err := p.Publish(context.Background(), "key", tt.body)

			require.NoError(t, err)
			assert.Equal(t, tt.body, capturedBody)
		})
	}
}
