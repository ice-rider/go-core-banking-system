package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Ack(tag uint64, multiple bool) error
	Nack(tag uint64, multiple bool, requeue bool) error
	Close() error
}

type Consumer struct {
	conn    *Connection
	queue   string
	handler func([]byte) error

	newChannel func() (ConsumerChannel, error)

	mu      sync.Mutex
	ch      ConsumerChannel
	stopCh  chan struct{}
	doneCh  chan struct{}
	stopped bool
}

func NewConsumer(conn *Connection, queue string, handler func([]byte) error) *Consumer {
	if handler == nil {
		panic("rabbitmq: handler must not be nil")
	}

	return &Consumer{
		conn:    conn,
		queue:   queue,
		handler: handler,
		newChannel: func() (ConsumerChannel, error) {
			ch, err := conn.Channel()
			if err != nil {
				return nil, err
			}
			return ch, nil
		},
		stopCh: make(chan struct{}),
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	ch, err := c.newChannel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	c.mu.Lock()
	c.ch = ch
	c.mu.Unlock()

	_, err = c.ch.QueueDeclare(c.queue, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	deliveries, err := c.ch.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.doneCh = make(chan struct{})
	go c.consume(ctx, deliveries)

	return nil
}

func (c *Consumer) consume(ctx context.Context, deliveries <-chan amqp.Delivery) {
	defer close(c.doneCh)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}
			if err := c.handler(delivery.Body); err != nil {
				_ = delivery.Nack(false, true)
			} else {
				_ = delivery.Ack(false)
			}
		}
	}
}

func (c *Consumer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return nil
	}

	c.stopped = true
	close(c.stopCh)

	if c.doneCh != nil {
		<-c.doneCh
	}

	if c.ch != nil {
		return c.ch.Close()
	}

	return nil
}

func (c *Consumer) Ack(delivery uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ch == nil {
		return fmt.Errorf("consumer not started")
	}

	return c.ch.Ack(delivery, false)
}

func (c *Consumer) Nack(delivery uint64, requeue bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ch == nil {
		return fmt.Errorf("consumer not started")
	}

	return c.ch.Nack(delivery, false, requeue)
}
