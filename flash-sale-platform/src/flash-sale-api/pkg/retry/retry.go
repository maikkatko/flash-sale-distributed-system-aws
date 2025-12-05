package retry

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
	Jitter      float64 // 0.0 to 1.0
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     5 * time.Second,
		Multiplier:  2.0,
		Jitter:      0.1,
	}
}

// RetryableError marks an error as retryable
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable checks if the error should be retried
func IsRetryable(err error) bool {
	var retryable *RetryableError
	return errors.As(err, &retryable)
}

// Do executes the function with retries
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Check context before attempting
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if error is not retryable
		if !IsRetryable(err) {
			return err
		}

		// Don't wait after last attempt
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// Calculate wait time with exponential backoff
		wait := cfg.InitialWait * time.Duration(math.Pow(cfg.Multiplier, float64(attempt)))
		if wait > cfg.MaxWait {
			wait = cfg.MaxWait
		}

		// Add jitter
		if cfg.Jitter > 0 {
			jitter := time.Duration(float64(wait) * cfg.Jitter * (rand.Float64()*2 - 1))
			wait += jitter
		}

		// Wait or context cancel
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}

	return lastErr
}

// DoWithResult executes a function that returns a value with retries
func DoWithResult[T any](ctx context.Context, cfg Config, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		val, err := fn()
		if err == nil {
			return val, nil
		}

		lastErr = err

		if !IsRetryable(err) {
			return result, err
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		wait := cfg.InitialWait * time.Duration(math.Pow(cfg.Multiplier, float64(attempt)))
		if wait > cfg.MaxWait {
			wait = cfg.MaxWait
		}

		if cfg.Jitter > 0 {
			jitter := time.Duration(float64(wait) * cfg.Jitter * (rand.Float64()*2 - 1))
			wait += jitter
		}

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(wait):
		}
	}

	return result, lastErr
}
