package circuit

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// Breaker implements the circuit breaker pattern
type Breaker struct {
	mu            sync.RWMutex
	name          string
	maxFailures   int
	resetTimeout  time.Duration
	halfOpenMax   int
	state         State
	failures      int
	successes     int
	lastFailure   time.Time
	halfOpenCount int
}

// Config holds circuit breaker configuration
type Config struct {
	Name         string
	MaxFailures  int           // failures before opening
	ResetTimeout time.Duration // time before trying half-open
	HalfOpenMax  int           // max requests in half-open state
}

// NewBreaker creates a new circuit breaker
func NewBreaker(cfg Config) *Breaker {
	if cfg.MaxFailures == 0 {
		cfg.MaxFailures = 5
	}
	if cfg.ResetTimeout == 0 {
		cfg.ResetTimeout = 30 * time.Second
	}
	if cfg.HalfOpenMax == 0 {
		cfg.HalfOpenMax = 3
	}

	return &Breaker{
		name:         cfg.Name,
		maxFailures:  cfg.MaxFailures,
		resetTimeout: cfg.ResetTimeout,
		halfOpenMax:  cfg.HalfOpenMax,
		state:        StateClosed,
	}
}

// Execute runs the given function if the circuit allows it
func (b *Breaker) Execute(fn func() error) error {
	if !b.AllowRequest() {
		return ErrCircuitOpen
	}

	err := fn()

	if err != nil {
		b.RecordFailure()
		return err
	}

	b.RecordSuccess()
	return nil
}

// AllowRequest checks if a request should be allowed
func (b *Breaker) AllowRequest() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if reset timeout has passed
		if time.Since(b.lastFailure) > b.resetTimeout {
			b.state = StateHalfOpen
			b.halfOpenCount = 0
			return true
		}
		return false

	case StateHalfOpen:
		// Allow limited requests in half-open state
		if b.halfOpenCount < b.halfOpenMax {
			b.halfOpenCount++
			return true
		}
		return false
	}

	return false
}

// RecordSuccess records a successful call
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		b.failures = 0

	case StateHalfOpen:
		b.successes++
		// If we've had enough successes, close the circuit
		if b.successes >= b.halfOpenMax {
			b.state = StateClosed
			b.failures = 0
			b.successes = 0
		}
	}
}

// RecordFailure records a failed call
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures++
	b.lastFailure = time.Now()

	switch b.state {
	case StateClosed:
		if b.failures >= b.maxFailures {
			b.state = StateOpen
		}

	case StateHalfOpen:
		// Any failure in half-open goes back to open
		b.state = StateOpen
		b.successes = 0
	}
}

// State returns the current state
func (b *Breaker) State() State {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

// Stats returns current circuit breaker statistics
func (b *Breaker) Stats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return map[string]interface{}{
		"name":         b.name,
		"state":        b.state.String(),
		"failures":     b.failures,
		"max_failures": b.maxFailures,
		"last_failure": b.lastFailure,
	}
}
