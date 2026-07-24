package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test-service")

	assert.Equal(t, "test-service", cfg.Name)
	assert.Equal(t, uint32(3), cfg.MaxRequests)
	assert.Equal(t, 30*time.Second, cfg.Interval)
	assert.Equal(t, 15*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.Retry.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, cfg.Retry.BaseDelay)
	assert.Equal(t, 2*time.Second, cfg.Retry.MaxDelay)
}

func TestNew(t *testing.T) {
	cfg := DefaultConfig("test")
	cb := New(cfg)

	require.NotNil(t, cb)
	assert.Equal(t, gobreaker.StateClosed, cb.State())
}

func TestExecute_Success(t *testing.T) {
	cb := New(DefaultConfig("test"))

	result, err := cb.Execute(context.Background(), time.Second, func(ctx context.Context) (any, error) {
		return "ok", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "ok", result)
}

func TestExecute_HandlerError(t *testing.T) {
	cb := New(DefaultConfig("test"))

	_, err := cb.Execute(context.Background(), time.Second, func(ctx context.Context) (any, error) {
		return nil, errors.New("service error")
	})

	require.Error(t, err)
	assert.EqualError(t, err, "service error")
}

func TestExecute_Timeout(t *testing.T) {
	cb := New(DefaultConfig("test"))

	_, err := cb.Execute(context.Background(), 50*time.Millisecond, func(ctx context.Context) (any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			return "ok", nil
		}
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestExecute_ConsecutiveFailures_TripsBreaker(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.ReadyToTrip = func(counts gobreaker.Counts) bool {
		return counts.ConsecutiveFailures >= 3
	}
	cb := New(cfg)

	for i := 0; i < 3; i++ {
		_, _ = cb.Execute(context.Background(), time.Second, func(ctx context.Context) (any, error) {
			return nil, errors.New("fail")
		})
	}

	assert.Equal(t, gobreaker.StateOpen, cb.State())

	_, err := cb.Execute(context.Background(), time.Second, func(ctx context.Context) (any, error) {
		return "should not run", nil
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, gobreaker.ErrOpenState)
}

func TestExecute_Retry_WithSuccess(t *testing.T) {
	cb := New(DefaultConfig("test"))

	var attempts atomic.Int32
	result, err := cb.ExecuteWithRetry(
		context.Background(),
		time.Second,
		func(ctx context.Context) (any, error) {
			if attempts.Add(1) < 3 {
				return nil, errors.New("temporary")
			}
			return "ok", nil
		},
		3,
		10*time.Millisecond,
	)

	require.NoError(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestExecute_Retry_AllFail(t *testing.T) {
	cb := New(DefaultConfig("test"))

	var attempts atomic.Int32
	_, err := cb.ExecuteWithRetry(
		context.Background(),
		time.Second,
		func(ctx context.Context) (any, error) {
			attempts.Add(1)
			return nil, errors.New("persistent")
		},
		3,
		10*time.Millisecond,
	)

	require.Error(t, err)
	assert.EqualError(t, err, "persistent")
	assert.Equal(t, int32(4), attempts.Load())
}

func TestExecute_Retry_ContextCancelled(t *testing.T) {
	cb := New(DefaultConfig("test"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := cb.ExecuteWithRetry(
		ctx,
		time.Second,
		func(ctx context.Context) (any, error) {
			return nil, errors.New("fail")
		},
		3,
		10*time.Millisecond,
	)

	require.Error(t, err)
}

func TestExecute_Retry_ExponentialBackoff(t *testing.T) {
	cb := New(DefaultConfig("test"))

	var delays []time.Duration
	var lastCall time.Time

	_, _ = cb.ExecuteWithRetry(
		context.Background(),
		time.Second,
		func(ctx context.Context) (any, error) {
			now := time.Now()
			if !lastCall.IsZero() {
				delays = append(delays, now.Sub(lastCall))
			}
			lastCall = now
			return nil, errors.New("fail")
		},
		3,
		50*time.Millisecond,
	)

	require.Len(t, delays, 3)
	for i := 1; i < len(delays); i++ {
		assert.GreaterOrEqual(t, delays[i], delays[i-1], "delays should increase")
	}
}

func TestExecute_SuccessResetsConsecutiveFailures(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.ReadyToTrip = func(counts gobreaker.Counts) bool {
		return counts.ConsecutiveFailures >= 3
	}
	cb := New(cfg)

	for i := 0; i < 2; i++ {
		_, _ = cb.Execute(context.Background(), time.Second, func(ctx context.Context) (any, error) {
			return nil, errors.New("fail")
		})
	}

	_, _ = cb.Execute(context.Background(), time.Second, func(ctx context.Context) (any, error) {
		return "ok", nil
	})

	assert.Equal(t, gobreaker.StateClosed, cb.State())
}
