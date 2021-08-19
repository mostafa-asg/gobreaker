package breaker

import (
	"sync"
	"sync/atomic"
	"time"
)

// StateReport returns the current state of the circut breaker
type StateReport struct {
	CurrentSate StateType
	Counts      Counts
	Generation  uint64
}

// CircuitBreaker is a state machine to prevent sending requests that are likely to fail.
type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool
	isSuccessful  func(err error) bool
	onStateChange func(name string, from StateType, to StateType)

	mutex        sync.Mutex
	currentState state
	generation   uint64
}

// NewCircuitBreaker returns a new CircuitBreaker configured with the given Settings.
func NewCircuitBreaker(st Settings) *CircuitBreaker {
	cb := new(CircuitBreaker)

	cb.name = st.Name
	cb.onStateChange = st.OnStateChange

	if st.MaxRequests == 0 {
		cb.maxRequests = 1
	} else {
		cb.maxRequests = st.MaxRequests
	}

	if st.Interval <= 0 {
		cb.interval = defaultInterval
	} else {
		cb.interval = st.Interval
	}

	if st.Timeout <= 0 {
		cb.timeout = defaultTimeout
	} else {
		cb.timeout = st.Timeout
	}

	if st.ReadyToTrip == nil {
		cb.readyToTrip = defaultReadyToTrip
	} else {
		cb.readyToTrip = st.ReadyToTrip
	}

	if st.IsSuccessful == nil {
		cb.isSuccessful = defaultIsSuccessful
	} else {
		cb.isSuccessful = st.IsSuccessful
	}

	closed = NewClosedState(cb)
	open = NewOpenState(cb)
	halfOpen = NewHalfOpenState(cb)

	cb.changeState(closed)

	return cb
}

// Name returns the name of the CircuitBreaker.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

func (cb *CircuitBreaker) Counts() Counts {
	return cb.currentState.getCounts()
}

// Execute runs the given request if the CircuitBreaker accepts it.
// Execute returns an error instantly if the CircuitBreaker rejects the request.
// Otherwise, Execute returns the result of the request.
// If a panic occurs in the request, the CircuitBreaker handles it as an error
// and causes the same panic again.
func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return cb.currentState.execute(req)
}

func (cb *CircuitBreaker) increaseGeneration() {
	atomic.AddUint64(&cb.generation, 1)
}

func (cb *CircuitBreaker) changeState(newState state) {
	prevState := cb.currentState

	if prevState != nil && prevState.getType() != newState.getType() {
		prevState.onLeave()
		cb.increaseGeneration()

		if cb.onStateChange != nil {
			cb.onStateChange(cb.name, prevState.getType(), newState.getType())
		}
	}

	cb.currentState = newState
	cb.currentState.onEnter()
}
