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
