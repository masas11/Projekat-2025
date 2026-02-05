package circuitbreaker

import (
	"log"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name           string
	maxFailures    int
	resetTimeout   time.Duration
	state          State
	failures       int
	lastFailTime   time.Time
	mu             sync.RWMutex
	onStateChange  func(name string, from, to State)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        Closed,
	}
}

// Execute runs the given function if the circuit breaker is closed
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Check if we should attempt to reset
	if cb.state == Open && time.Since(cb.lastFailTime) > cb.resetTimeout {
		cb.setState(HalfOpen)
	}

	if cb.state == Open {
		return &CircuitBreakerError{Message: "circuit breaker is open"}
	}

	err := fn()
	
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// onSuccess resets failure count and closes circuit
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	if cb.state != Closed {
		cb.setState(Closed)
	}
}

// onFailure increments failure count and opens circuit if threshold reached
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.setState(Open)
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState State) {
	oldState := cb.state
	cb.state = newState
	
	log.Printf("Circuit breaker '%s' changed from %v to %v (failures: %d)", 
		cb.name, oldState, newState, cb.failures)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, oldState, newState)
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// CircuitBreakerError is returned when the circuit breaker is open
type CircuitBreakerError struct {
	Message string
}

func (e *CircuitBreakerError) Error() string {
	return e.Message
}
