package breaker

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type StateChange struct {
	name string
	from StateType
	to   StateType
}

var stateChange StateChange

func fail(cb *CircuitBreaker) error {
	_, err := cb.Execute(func() (interface{}, error) { return nil, fmt.Errorf("fail") })
	return err
}

func succeed(cb *CircuitBreaker) error {
	_, err := cb.Execute(func() (interface{}, error) { return nil, nil })
	return err
}

func succeedLater(cb *CircuitBreaker, delay time.Duration) <-chan error {
	ch := make(chan error)
	go func() {
		_, err := cb.Execute(func() (interface{}, error) {
			time.Sleep(delay)
			return nil, nil
		})
		ch <- err
	}()
	return ch
}

func causePanic(cb *CircuitBreaker) error {
	_, err := cb.Execute(func() (interface{}, error) { panic("oops"); return nil, nil })
	return err
}

func newCustom() *CircuitBreaker {
	var customSt Settings
	customSt.Name = "cb"
	customSt.MaxRequests = 3
	customSt.Interval = 3 * time.Second
	customSt.Timeout = 2 * time.Second
	customSt.ReadyToTrip = func(counts Counts) bool {
		numReqs := counts.Requests
		failureRatio := float64(counts.TotalFailures) / float64(numReqs)
		return numReqs >= 3 && failureRatio >= 0.6
	}
	customSt.OnStateChange = func(name string, from StateType, to StateType) {
		stateChange = StateChange{name, from, to}
	}

	return NewCircuitBreaker(customSt)
}

func newNegativeDurationCB() *CircuitBreaker {
	var negativeSt Settings
	negativeSt.Name = "ncb"
	negativeSt.Interval = time.Duration(-30) * time.Second
	negativeSt.Timeout = time.Duration(-90) * time.Second

	return NewCircuitBreaker(negativeSt)
}

func TestNewCircuitBreaker(t *testing.T) {
	defaultCB := NewCircuitBreaker(Settings{})
	assert.Equal(t, "", defaultCB.name)
	assert.Equal(t, uint32(1), defaultCB.maxRequests)
	assert.Equal(t, time.Duration(0), defaultCB.interval)
	assert.Equal(t, time.Duration(60)*time.Second, defaultCB.timeout)
	assert.NotNil(t, defaultCB.readyToTrip)
	assert.Nil(t, defaultCB.onStateChange)
	assert.Equal(t, Closed, defaultCB.currentState.getType())
	assert.Equal(t, Counts{}, defaultCB.Counts())

	customCB := newCustom()
	assert.Equal(t, "cb", customCB.name)
	assert.Equal(t, uint32(3), customCB.maxRequests)
	assert.Equal(t, 3*time.Second, customCB.interval)
	assert.Equal(t, 2*time.Second, customCB.timeout)
	assert.NotNil(t, customCB.readyToTrip)
	assert.NotNil(t, customCB.onStateChange)
	assert.Equal(t, Closed, customCB.currentState.getType())
	assert.Equal(t, Counts{}, customCB.Counts())

	negativeDurationCB := newNegativeDurationCB()
	assert.Equal(t, "ncb", negativeDurationCB.name)
	assert.Equal(t, uint32(1), negativeDurationCB.maxRequests)
	assert.Equal(t, time.Duration(0)*time.Second, negativeDurationCB.interval)
	assert.Equal(t, time.Duration(60)*time.Second, negativeDurationCB.timeout)
	assert.NotNil(t, negativeDurationCB.readyToTrip)
	assert.Nil(t, negativeDurationCB.onStateChange)
	assert.Equal(t, Closed, negativeDurationCB.currentState.getType())
	assert.Equal(t, Counts{}, negativeDurationCB.Counts())
}

func TestDefaultCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(Settings{
		Timeout: 2 * time.Second,
	})
	assert.Equal(t, "", cb.Name())

	for i := 0; i < 5; i++ {
		fail(cb)
	}
	assert.Equal(t, Closed, cb.currentState.getType())
	assert.Equal(t, NewCounts(5, 0, 5, 0, 5), cb.Counts())

	succeed(cb)
	assert.Equal(t, Closed, cb.currentState.getType())
	assert.Equal(t, NewCounts(6, 1, 5, 1, 0), cb.Counts())

	fail(cb)
	assert.Equal(t, Closed, cb.currentState.getType())
	assert.Equal(t, NewCounts(7, 1, 6, 0, 1), cb.Counts())

	// StateClosed to StateOpen
	for i := 0; i < 5; i++ {
		fail(cb) // 6 consecutive failures
	}
	assert.Equal(t, Open, cb.currentState.getType())
	assert.Equal(t, NewCounts(0, 0, 0, 0, 0), cb.Counts())

	assert.Equal(t, ErrOpenState, succeed(cb))
	assert.Equal(t, ErrOpenState, fail(cb))
	assert.Equal(t, NewCounts(0, 0, 0, 0, 0), cb.Counts())

	time.Sleep(1 * time.Second)
	assert.Equal(t, Open, cb.currentState.getType())

	// StateOpen to StateHalfOpen
	time.Sleep(1100 * time.Millisecond)
	assert.Equal(t, HalfOpen, cb.currentState.getType())

	// StateHalfOpen to StateOpen
	fail(cb)
	assert.Equal(t, Open, cb.currentState.getType())
	assert.Equal(t, NewCounts(0, 0, 0, 0, 0), cb.Counts())

	// StateOpen to StateHalfOpen
	time.Sleep(2100 * time.Millisecond)
	assert.Equal(t, HalfOpen, cb.currentState.getType())

	// StateHalfOpen to StateClosed
	succeed(cb)
	assert.Equal(t, Closed, cb.currentState.getType())
	assert.Equal(t, NewCounts(0, 0, 0, 0, 0), cb.Counts())
}

func TestCustomCircuitBreaker(t *testing.T) {
	customCB := newCustom()
	assert.Equal(t, "cb", customCB.Name())

	for i := 0; i < 5; i++ {
		succeed(customCB)
		fail(customCB)
	}
	assert.Equal(t, Closed, customCB.currentState.getType())
	assert.Equal(t, NewCounts(10, 5, 5, 0, 1), customCB.Counts())

	time.Sleep(2 * time.Second)
	succeed(customCB)
	assert.Equal(t, Closed, customCB.currentState.getType())
	assert.Equal(t, NewCounts(11, 6, 5, 1, 0), customCB.Counts())

	time.Sleep(1100 * time.Millisecond) // over Interval
	fail(customCB)
	assert.Equal(t, Closed, customCB.currentState.getType())
	assert.Equal(t, NewCounts(1, 0, 1, 0, 1), customCB.Counts())

	// StateClosed to StateOpen
	succeed(customCB)
	fail(customCB) // failure ratio: 2/3 >= 0.6
	assert.Equal(t, Open, customCB.currentState.getType())
	assert.Equal(t, NewCounts(0, 0, 0, 0, 0), customCB.Counts())
	assert.Equal(t, StateChange{"cb", Closed, Open}, stateChange)

	// StateOpen to StateHalfOpen
	time.Sleep(2100 * time.Millisecond)
	assert.Equal(t, HalfOpen, customCB.currentState.getType())
	assert.Equal(t, StateChange{"cb", Open, HalfOpen}, stateChange)

	succeed(customCB)
	succeed(customCB)
	assert.Equal(t, HalfOpen, customCB.currentState.getType())
	assert.Equal(t, NewCounts(2, 2, 0, 2, 0), customCB.Counts())

	//StateHalfOpen to StateClosed
	succeedLater(customCB, 100*time.Millisecond) // 3 consecutive successes
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, NewCounts(3, 2, 0, 2, 0), customCB.Counts())
	succeed(customCB) // over MaxRequests
	assert.Equal(t, Closed, customCB.currentState.getType())
	assert.Equal(t, NewCounts(0, 0, 0, 0, 0), customCB.Counts())
	assert.Equal(t, StateChange{"cb", HalfOpen, Closed}, stateChange)
}

func TestPanicInRequest(t *testing.T) {
	cb := NewCircuitBreaker(Settings{})
	assert.Panics(t, func() { causePanic(cb) })
	assert.Equal(t, NewCounts(1, 0, 1, 0, 1), cb.Counts())
}

func TestGeneration(t *testing.T) {
	cb := newCustom()
	succeed(cb)
	ch := succeedLater(cb, 3500*time.Millisecond)
	time.Sleep(2500 * time.Millisecond)

	assert.Equal(t, NewCounts(2, 1, 0, 1, 0), cb.Counts())

	time.Sleep(time.Duration(600) * time.Millisecond) // over Interval
	assert.Equal(t, Closed, cb.currentState.getType())
	assert.Equal(t, NewCounts(0, 0, 0, 0, 0), cb.Counts())

	// the request from the previous generation has no effect on customCB.counts
	assert.Nil(t, <-ch)
	assert.Equal(t, NewCounts(0, 0, 0, 0, 0), cb.Counts())
}
