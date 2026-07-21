package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherChannel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Close() error
}

type Publisher struct {
	conn     *Connection
	exchange string

	newChannel func() (PublisherChannel, error)

	mu     sync.Mutex
	ch     PublisherChannel
	closed bool
}

func NewPublisher(conn *Connection, exchange string) *Publisher {
	return &Publisher{
		conn:     conn,
		exchange: exchange,
		newChannel: func() (PublisherChannel, error) {
			return conn.Channel()
		},
	}
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("publisher is closed")
	}

	if p.ch == nil {
		ch, err := p.newChannel()
		if err != nil {
			return fmt.Errorf("failed to open channel: %w", err)
		}
		p.ch = ch
	}

	return p.ch.Publish(
		p.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	if p.ch != nil {
		if err := p.ch.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	return nil
}
