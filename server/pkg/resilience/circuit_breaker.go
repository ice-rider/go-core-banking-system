package resilience

import (
	"context"
	"log/slog"
	"time"

	"github.com/sony/gobreaker/v2"
)

// Config holds circuit breaker + retry configuration.
type Config struct {
	Name          string
	MaxRequests   uint32
	Interval      time.Duration
	Timeout       time.Duration
	ReadyToTrip   func(counts gobreaker.Counts) bool
	OnStateChange func(name string, from, to gobreaker.State)
	Retry         RetryConfig
}

// RetryConfig holds retry configuration.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultConfig returns sensible defaults for a service circuit breaker.
func DefaultConfig(name string) Config {
	return Config{
		Name:        name,
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			slog.Warn("circuit breaker state change",
				"name", name,
				"from", from.String(),
				"to", to.String(),
			)
		},
		Retry: RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    2 * time.Second,
		},
	}
}

// CircuitBreaker wraps gobreaker with retry and timeout.
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker[any]
}

// New creates a new CircuitBreaker with the given config.
func New(cfg Config) *CircuitBreaker {
	cbSettings := gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: cfg.ReadyToTrip,
		OnStateChange: func(name string, from, to gobreaker.State) {
			if cfg.OnStateChange != nil {
				cfg.OnStateChange(name, from, to)
			}
		},
	}

	return &CircuitBreaker{
		cb: gobreaker.NewCircuitBreaker[any](cbSettings),
	}
}

// Execute runs the given function with circuit breaker + retry + timeout.
func (c *CircuitBreaker) Execute(ctx context.Context, timeout time.Duration, fn func(ctx context.Context) (any, error)) (any, error) {
	return c.ExecuteWithRetry(ctx, timeout, fn, 0, 0)
}

// ExecuteWithRetry runs the function with circuit breaker, retry, and timeout.
func (c *CircuitBreaker) ExecuteWithRetry(ctx context.Context, timeout time.Duration, fn func(ctx context.Context) (any, error), maxAttempts int, baseDelay time.Duration) (any, error) {
	var lastErr error

	for attempt := 0; attempt <= maxAttempts; attempt++ {
		result, err := c.cb.Execute(func() (any, error) {
			callCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return fn(callCtx)
		})

		if err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests {
			return nil, err
		}

		if err == nil {
			return result, nil
		}

		lastErr = err

		if attempt < maxAttempts {
			delay := baseDelay * time.Duration(1<<uint(attempt))
			if delay > 2*time.Second {
				delay = 2 * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return nil, lastErr
}

// State returns the current circuit breaker state.
func (c *CircuitBreaker) State() gobreaker.State {
	return c.cb.State()
}
