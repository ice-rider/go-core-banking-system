package rabbitmq

import (
	"context"
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConsumerChannel struct {
	queueDeclareFunc func(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	consumeFunc      func(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	ackFunc          func(tag uint64, multiple bool) error
	nackFunc         func(tag uint64, multiple bool, requeue bool) error
	closeFunc        func() error
}

func (m *mockConsumerChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	if m.queueDeclareFunc != nil {
		return m.queueDeclareFunc(name, durable, autoDelete, exclusive, noWait, args)
	}
	return amqp.Queue{Name: name}, nil
}

func (m *mockConsumerChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if m.consumeFunc != nil {
		return m.consumeFunc(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	}
	ch := make(chan amqp.Delivery)
	close(ch)
	return ch, nil
}

func (m *mockConsumerChannel) Ack(tag uint64, multiple bool) error {
	if m.ackFunc != nil {
		return m.ackFunc(tag, multiple)
	}
	return nil
}

func (m *mockConsumerChannel) Nack(tag uint64, multiple bool, requeue bool) error {
	if m.nackFunc != nil {
		return m.nackFunc(tag, multiple, requeue)
	}
	return nil
}

func (m *mockConsumerChannel) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestNewConsumer_Success(t *testing.T) {
	handler := func(data []byte) error { return nil }

	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := NewConsumer(c, "test-queue", handler)

	require.NotNil(t, consumer)
	assert.False(t, consumer.stopped)
}

func TestNewConsumer_NilHandler(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	require.Panics(t, func() {
		NewConsumer(c, "test-queue", nil)
	})
}

func TestStart_Success(t *testing.T) {
	mockCh := &mockConsumerChannel{}

	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		newChannel: func() (ConsumerChannel, error) {
			return mockCh, nil
		},
		stopCh: make(chan struct{}),
	}

	err = consumer.Start(context.Background())

	require.NoError(t, err)
	require.NoError(t, consumer.Stop())
}

func TestStart_ContextCancelled(t *testing.T) {
	mockCh := &mockConsumerChannel{}

	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		newChannel: func() (ConsumerChannel, error) {
			return mockCh, nil
		},
		stopCh: make(chan struct{}),
	}

	err = consumer.Start(ctx)
	require.NoError(t, err)

	cancel()
	require.NoError(t, consumer.Stop())
}

func TestStart_ConnectionClosed(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		newChannel: func() (ConsumerChannel, error) {
			return nil, errors.New("channel error")
		},
		stopCh: make(chan struct{}),
	}

	err = consumer.Start(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open channel")
}

func TestStart_QueueDeclareError(t *testing.T) {
	mockCh := &mockConsumerChannel{
		queueDeclareFunc: func(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
			return amqp.Queue{}, errors.New("queue declare error")
		},
	}

	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		newChannel: func() (ConsumerChannel, error) {
			return mockCh, nil
		},
		stopCh: make(chan struct{}),
	}

	err = consumer.Start(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to declare queue")
}

func TestStart_ConsumeError(t *testing.T) {
	mockCh := &mockConsumerChannel{
		consumeFunc: func(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
			return nil, errors.New("consume error")
		},
	}

	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		newChannel: func() (ConsumerChannel, error) {
			return mockCh, nil
		},
		stopCh: make(chan struct{}),
	}

	err = consumer.Start(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start consuming")
}

func TestStop_Success(t *testing.T) {
	mockCh := &mockConsumerChannel{}

	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		newChannel: func() (ConsumerChannel, error) {
			return mockCh, nil
		},
		stopCh: make(chan struct{}),
	}

	err = consumer.Start(context.Background())
	require.NoError(t, err)

	err = consumer.Stop()

	require.NoError(t, err)
	assert.True(t, consumer.stopped)
}

func TestStop_AlreadyStopped(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		stopCh:  make(chan struct{}),
	}

	err = consumer.Stop()
	require.NoError(t, err)

	err = consumer.Stop()
	require.NoError(t, err)
}

func TestAck_Success(t *testing.T) {
	mockCh := &mockConsumerChannel{}

	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		newChannel: func() (ConsumerChannel, error) {
			return mockCh, nil
		},
		stopCh: make(chan struct{}),
	}

	err = consumer.Start(context.Background())
	require.NoError(t, err)
	defer func() { _ = consumer.Stop() }()

	err = consumer.Ack(1)

	require.NoError(t, err)
}

func TestAck_InvalidDelivery(t *testing.T) {
	t.Run("consumer not started", func(t *testing.T) {
		conn := &mockConnection{}
		restore := setupMockDialer(func(url string) (AMQPConnection, error) {
			return conn, nil
		})
		defer restore()

		c, err := NewConnection("amqp://guest:guest@localhost:5672/")
		require.NoError(t, err)

		consumer := NewConsumer(c, "test-queue", func(data []byte) error { return nil })

		err = consumer.Ack(1)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "consumer not started")
	})

	t.Run("ack channel error", func(t *testing.T) {
		mockCh := &mockConsumerChannel{
			ackFunc: func(tag uint64, multiple bool) error {
				return errors.New("unknown delivery tag")
			},
		}

		conn := &mockConnection{}
		restore := setupMockDialer(func(url string) (AMQPConnection, error) {
			return conn, nil
		})
		defer restore()

		c, err := NewConnection("amqp://guest:guest@localhost:5672/")
		require.NoError(t, err)

		consumer := &Consumer{
			conn:    c,
			queue:   "test-queue",
			handler: func(data []byte) error { return nil },
			newChannel: func() (ConsumerChannel, error) {
				return mockCh, nil
			},
			stopCh: make(chan struct{}),
		}

		err = consumer.Start(context.Background())
		require.NoError(t, err)
		defer func() { _ = consumer.Stop() }()

		err = consumer.Ack(999)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown delivery tag")
	})
}

func TestNack_Success(t *testing.T) {
	mockCh := &mockConsumerChannel{}

	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		newChannel: func() (ConsumerChannel, error) {
			return mockCh, nil
		},
		stopCh: make(chan struct{}),
	}

	err = consumer.Start(context.Background())
	require.NoError(t, err)
	defer func() { _ = consumer.Stop() }()

	err = consumer.Nack(1, true)

	require.NoError(t, err)
}

func TestNack_WithoutRequeue(t *testing.T) {
	mockCh := &mockConsumerChannel{
		nackFunc: func(tag uint64, multiple bool, requeue bool) error {
			assert.Equal(t, uint64(5), tag)
			assert.False(t, multiple)
			assert.False(t, requeue)
			return nil
		},
	}

	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := &Consumer{
		conn:    c,
		queue:   "test-queue",
		handler: func(data []byte) error { return nil },
		newChannel: func() (ConsumerChannel, error) {
			return mockCh, nil
		},
		stopCh: make(chan struct{}),
	}

	err = consumer.Start(context.Background())
	require.NoError(t, err)
	defer func() { _ = consumer.Stop() }()

	err = consumer.Nack(5, false)

	require.NoError(t, err)
}

func TestNack_NotStarted(t *testing.T) {
	conn := &mockConnection{}
	restore := setupMockDialer(func(url string) (AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	c, err := NewConnection("amqp://guest:guest@localhost:5672/")
	require.NoError(t, err)

	consumer := NewConsumer(c, "test-queue", func(data []byte) error { return nil })

	err = consumer.Nack(1, true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "consumer not started")
}
