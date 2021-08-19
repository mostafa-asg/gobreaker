package breaker

import "fmt"

// StateType is a type that represents a state of CircuitBreaker.
type StateType int

// These constants are states of CircuitBreaker.
const (
	Closed StateType = iota
	HalfOpen
	Open
)

// String implements stringer interface.
func (s StateType) String() string {
	switch s {
	case Closed:
		return "closed"
	case HalfOpen:
		return "half-open"
	case Open:
		return "open"
	default:
		return fmt.Sprintf("unknown state: %d", s)
	}
}
