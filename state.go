package breaker

type state interface {
	execute(func() (interface{}, error)) (interface{}, error)
	getCounts() Counts
	getType() StateType

	// for allocation resources
	onEnter()
	// for releasing resources
	onLeave()
}

// These are states of CircuitBreaker.
var (
	closed   *closedState
	halfOpen *halfOpenState
	open     *openState
)
