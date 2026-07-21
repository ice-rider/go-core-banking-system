package rabbitmq

import (
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConnection struct {
	channelFunc func() (*amqp.Channel, error)
	closeFunc   func() error
	isClosed    bool
}

func (m *mockConnection) Channel() (*amqp.Channel, error) {
	if m.channelFunc != nil {
		return m.channelFunc()
	}
	return nil, nil
}

func (m *mockConnection) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	m.isClosed = true
	return nil
}

func (m *mockConnection) IsClosed() bool {
	return m.isClosed
}

type mockDialer struct {
	dialFunc func(url string) (AMQPConnection, error)
}

func (m *mockDialer) Dial(url string) (AMQPConnection, error) {
	if m.dialFunc != nil {
		return m.dialFunc(url)
	}
	return nil, nil
}

func setupMockDialer(dialFunc func(string) (AMQPConnection, error)) func() {
	orig := NewDialer
	NewDialer = func() Dialer { return &mockDialer{dialFunc: dialFunc} }
	return func() { NewDialer = orig }
}

func TestNewConnection_Success(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		assert.Equal(t, "amqp://guest:guest@localhost:5672/", url)
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")

	require.NoError(t, err)
	require.NotNil(t, c)
	assert.True(t, c.IsConnected())
}

func TestNewConnection_InvalidURL(t *testing.T) {
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return nil, errors.New("dial: invalid URL")
	})
	defer restore()

	c, err := NewConnection("not-a-url")

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "failed to connect to RabbitMQ")
}

func TestNewConnection_ConnectionClosed(t *testing.T) {
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return nil, errors.New("connection closed by server")
	})
	defer restore()

	c, err := NewConnection("amqp://closed-host:5672/")

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "connection closed by server")
}

func TestClose_Success(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	err = c.Close()

	require.NoError(t, err)
	assert.False(t, c.IsConnected())
}

func TestClose_AlreadyClosed(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	require.NoError(t, c.Close())

	err = c.Close()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already closed")
}

func TestIsConnected_True(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	assert.True(t, c.IsConnected())
}

func TestIsConnected_False(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)
	require.NoError(t, c.Close())

	assert.False(t, c.IsConnected())
}

func TestChannel_Success(t *testing.T) {
	conn := &mockConnection{
		channelFunc: func() (*amqp.Channel, error) {
			return &amqp.Channel{}, nil
		},
	}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	ch, err := c.Channel()

	require.NoError(t, err)
	assert.NotNil(t, ch)
}

func TestChannel_ConnectionClosed(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)
	require.NoError(t, c.Close())

	ch, err := c.Channel()

	require.Error(t, err)
	assert.Nil(t, ch)
	assert.Contains(t, err.Error(), "connection is closed")
}

func TestChannel_AMQPError(t *testing.T) {
	conn := &mockConnection{
		channelFunc: func() (*amqp.Channel, error) {
			return nil, &amqp.Error{Code: 403, Reason: "access refused"}
		},
	}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	ch, err := c.Channel()

	require.Error(t, err)
	assert.Nil(t, ch)
}

func TestClose_AMQPError(t *testing.T) {
	conn := &mockConnection{
		closeFunc: func() error {
			return errors.New("unexpected close error")
		},
	}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	err = c.Close()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected close error")
	assert.False(t, c.IsConnected())
}
