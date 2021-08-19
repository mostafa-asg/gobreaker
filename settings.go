package breaker

import "time"

// Settings configures CircuitBreaker:
type Settings struct {
	// Name is the name of the CircuitBreaker.
	Name string

	// MaxRequests is the maximum number of requests allowed to pass through
	// when the CircuitBreaker is half-open.
	// If MaxRequests is 0, the CircuitBreaker allows only 1 request.
	MaxRequests uint32

	// Interval is the cyclic period of the closed state
	// for the CircuitBreaker to clear the internal Counts.
	// If Interval is less than or equal to 0, the CircuitBreaker doesn't clear internal Counts during the closed state.
	Interval time.Duration

	// Timeout is the period of the open state,
	// after which the state of the CircuitBreaker becomes half-open.
	// If Timeout is less than or equal to 0, the timeout value of the CircuitBreaker is set to 60 seconds.
	Timeout time.Duration

	// ReadyToTrip is called with a copy of Counts whenever a request fails in the closed state.
	// If ReadyToTrip returns true, the CircuitBreaker will be placed into the open state.
	// If ReadyToTrip is nil, default ReadyToTrip is used.
	// Default ReadyToTrip returns true when the number of consecutive failures is more than 5.
	ReadyToTrip func(counts Counts) bool

	// OnStateChange is called whenever the state of the CircuitBreaker changes.
	OnStateChange func(name string, from StateType, to StateType)

	// IsSuccessful is called with the error returned from the request, if not nil.
	// If IsSuccessful returns false, the error is considered a failure, and is counted towards tripping the circuit breaker.
	// If IsSuccessful returns true, the error will be returned to the caller without tripping the circuit breaker.
	// If IsSuccessful is nil, default IsSuccessful is used, which returns false for all non-nil errors.
	IsSuccessful func(err error) bool
}
