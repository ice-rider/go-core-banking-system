package rabbitmq

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPConnection interface {
	Channel() (*amqp.Channel, error)
	Close() error
	IsClosed() bool
}

type Dialer interface {
	Dial(url string) (AMQPConnection, error)
}

type defaultDialer struct{}

func (d *defaultDialer) Dial(url string) (AMQPConnection, error) {
	return amqp.Dial(url)
}

var NewDialer = func() Dialer { return &defaultDialer{} }

type Connection struct {
	mu     sync.RWMutex
	conn   AMQPConnection
	closed bool
}

func NewConnection(url string) (*Connection, error) {
	conn, err := NewDialer().Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &Connection{
		conn: conn,
	}, nil
}

func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection is already closed")
	}

	err := c.conn.Close()
	c.conn = nil
	c.closed = true
	return err
}

func (c *Connection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.conn != nil
}

func (c *Connection) Channel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return nil, fmt.Errorf("connection is closed")
	}

	return c.conn.Channel()
}
