package rabbitmq

import (
	"context"
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockChannel struct {
	exchangeDeclareFunc func(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	queueDeclareFunc    func(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	queueBindFunc       func(name, key, exchange string, noWait bool, args amqp.Table) error
}

func (m *mockChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	if m.exchangeDeclareFunc != nil {
		return m.exchangeDeclareFunc(name, kind, durable, autoDelete, internal, noWait, args)
	}
	return nil
}

func (m *mockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	if m.queueDeclareFunc != nil {
		return m.queueDeclareFunc(name, durable, autoDelete, exclusive, noWait, args)
	}
	return amqp.Queue{Name: name}, nil
}

func (m *mockChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	if m.queueBindFunc != nil {
		return m.queueBindFunc(name, key, exchange, noWait, args)
	}
	return nil
}

func TestDefaultDLQConfig(t *testing.T) {
	cfg := DefaultDLQConfig("orders")

	assert.Equal(t, "orders.dlq", cfg.Exchange)
	assert.Equal(t, "orders.dlq", cfg.Queue)
	assert.Equal(t, "orders.dlq", cfg.RoutingKey)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 1*time.Second, cfg.RetryDelay)
}

func TestNewDLQPublisher(t *testing.T) {
	c := createTestConnection(t)
	cfg := DefaultDLQConfig("test-queue")

	dlq := NewDLQPublisher(c, cfg)

	require.NotNil(t, dlq)
	assert.Equal(t, cfg.Exchange, dlq.publisher.exchange)
}

func TestSetupExchange_Success(t *testing.T) {
	mockCh := &mockChannel{}
	cfg := DefaultDLQConfig("test-queue")
	dlq := &DLQPublisher{config: cfg}

	err := dlq.SetupExchange(mockCh)

	require.NoError(t, err)
}

func TestSetupExchange_ExchangeDeclareError(t *testing.T) {
	mockCh := &mockChannel{
		exchangeDeclareFunc: func(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
			return errors.New("exchange declare failed")
		},
	}
	cfg := DefaultDLQConfig("test-queue")
	dlq := &DLQPublisher{config: cfg}

	err := dlq.SetupExchange(mockCh)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to declare DLQ exchange")
}

func TestSetupExchange_QueueDeclareError(t *testing.T) {
	mockCh := &mockChannel{
		queueDeclareFunc: func(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
			return amqp.Queue{}, errors.New("queue declare failed")
		},
	}
	cfg := DefaultDLQConfig("test-queue")
	dlq := &DLQPublisher{config: cfg}

	err := dlq.SetupExchange(mockCh)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to declare DLQ queue")
}

func TestSetupExchange_QueueBindError(t *testing.T) {
	mockCh := &mockChannel{
		queueBindFunc: func(name, key, exchange string, noWait bool, args amqp.Table) error {
			return errors.New("queue bind failed")
		},
	}
	cfg := DefaultDLQConfig("test-queue")
	dlq := &DLQPublisher{config: cfg}

	err := dlq.SetupExchange(mockCh)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to bind DLQ queue")
}

func TestSetupExchange_VerifiesArgs(t *testing.T) {
	var declaredExchange, declaredQueue, declaredRoutingKey string
	mockCh := &mockChannel{
		exchangeDeclareFunc: func(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
			declaredExchange = name
			assert.Equal(t, "direct", kind)
			assert.True(t, durable)
			return nil
		},
		queueDeclareFunc: func(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
			declaredQueue = name
			assert.True(t, durable)
			return amqp.Queue{Name: name}, nil
		},
		queueBindFunc: func(name, key, exchange string, noWait bool, args amqp.Table) error {
			declaredRoutingKey = key
			assert.Equal(t, exchange, declaredExchange)
			return nil
		},
	}
	cfg := DefaultDLQConfig("transactions")
	dlq := &DLQPublisher{config: cfg}

	err := dlq.SetupExchange(mockCh)

	require.NoError(t, err)
	assert.Equal(t, "transactions.dlq", declaredExchange)
	assert.Equal(t, "transactions.dlq", declaredQueue)
	assert.Equal(t, "transactions.dlq", declaredRoutingKey)
}

func TestDLQPublishWithRetry_Success(t *testing.T) {
	mockCh := &mockPublisherChannel{}

	c := createTestConnection(t)
	cfg := DefaultDLQConfig("test-queue")
	cfg.MaxRetries = 2
	cfg.RetryDelay = 1 * time.Millisecond

	dlq := &DLQPublisher{
		publisher: &Publisher{
			conn:     c,
			exchange: cfg.Exchange,
			newChannel: func() (PublisherChannel, error) {
				return mockCh, nil
			},
		},
		config: cfg,
	}

	publishCount := 0
	mockCh.publishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		publishCount++
		if publishCount == 1 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := dlq.PublishWithRetry(context.Background(), "test-key", []byte("data"))

	require.NoError(t, err)
	assert.Equal(t, 2, publishCount)
}

func TestDLQPublishWithRetry_AllRetriesFailed(t *testing.T) {
	mockCh := &mockPublisherChannel{}

	c := createTestConnection(t)
	cfg := DefaultDLQConfig("test-queue")
	cfg.MaxRetries = 2
	cfg.RetryDelay = 1 * time.Millisecond

	dlq := &DLQPublisher{
		publisher: &Publisher{
			conn:     c,
			exchange: cfg.Exchange,
			newChannel: func() (PublisherChannel, error) {
				return mockCh, nil
			},
		},
		config: cfg,
	}

	mockCh.publishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		return errors.New("persistent failure")
	}

	err := dlq.PublishWithRetry(context.Background(), "test-key", []byte("data"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed after 2 retries")
}

func TestDLQPublishWithRetry_ZeroRetries(t *testing.T) {
	mockCh := &mockPublisherChannel{}

	c := createTestConnection(t)
	cfg := DefaultDLQConfig("test-queue")
	cfg.MaxRetries = 0

	dlq := &DLQPublisher{
		publisher: &Publisher{
			conn:     c,
			exchange: cfg.Exchange,
			newChannel: func() (PublisherChannel, error) {
				return mockCh, nil
			},
		},
		config: cfg,
	}

	mockCh.publishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		return errors.New("failure")
	}

	err := dlq.PublishWithRetry(context.Background(), "test-key", []byte("data"))

	require.Error(t, err)
}
