package breaker

import (
	"errors"
	"time"
)

var ErrOpenState = errors.New("circuit breaker is open")

type openState struct {
	counts      *Counts
	ticker      *time.Ticker
	stateLeaved chan interface{}
	cb          *CircuitBreaker
}

func NewOpenState(cb *CircuitBreaker) *openState {
	s := &openState{
		cb:     cb,
		counts: &Counts{},
	}

	return s
}

func (s *openState) onEnter() {
	cb := s.cb
	s.ticker = time.NewTicker(cb.timeout)
	s.stateLeaved = make(chan interface{})
	go func() {
		for {
			select {
			case <-s.ticker.C:
				if cb.currentState.getType() == Open {
					cb.changeState(halfOpen)
				}
			case <-s.stateLeaved:
				return
			}
		}
	}()
}

func (s *openState) onLeave() {
	s.counts.clear()
	s.ticker.Stop()
	close(s.stateLeaved)
}

func (s *openState) getType() StateType {
	return Open
}

func (s *openState) execute(req func() (interface{}, error)) (interface{}, error) {
	return nil, ErrOpenState
}

func (s *openState) getCounts() Counts {
	return *s.counts
}
