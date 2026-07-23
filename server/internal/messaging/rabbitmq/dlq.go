package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DLQConfig struct {
	Exchange   string
	Queue      string
	RoutingKey string
	MaxRetries int
	RetryDelay time.Duration
}

func DefaultDLQConfig(queue string) DLQConfig {
	return DLQConfig{
		Exchange:   fmt.Sprintf("%s.dlq", queue),
		Queue:      fmt.Sprintf("%s.dlq", queue),
		RoutingKey: fmt.Sprintf("%s.dlq", queue),
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

type DLQPublisher struct {
	publisher *Publisher
	config    DLQConfig
}

func NewDLQPublisher(conn *Connection, config DLQConfig) *DLQPublisher {
	return &DLQPublisher{
		publisher: NewPublisher(conn, config.Exchange),
		config:    config,
	}
}

func (d *DLQPublisher) SetupExchange(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(
		d.config.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare DLQ exchange: %w", err)
	}

	if _, err := channel.QueueDeclare(
		d.config.Queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare DLQ queue: %w", err)
	}

	if err := channel.QueueBind(
		d.config.Queue,
		d.config.RoutingKey,
		d.config.Exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind DLQ queue: %w", err)
	}

	return nil
}

func (d *DLQPublisher) PublishWithRetry(ctx context.Context, routingKey string, body []byte) error {
	var lastErr error

	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		if err := d.publisher.Publish(ctx, routingKey, body); err != nil {
			lastErr = err
			if attempt < d.config.MaxRetries {
				time.Sleep(d.config.RetryDelay * time.Duration(attempt+1))
				continue
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("failed after %d retries: %w", d.config.MaxRetries, lastErr)
}
