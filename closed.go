package breaker

import (
	"time"
)

type closedState struct {
	counts      *Counts
	ticker      *time.Ticker
	stateLeaved chan interface{}
	cb          *CircuitBreaker
}

func NewClosedState(cb *CircuitBreaker) *closedState {
	c := &closedState{
		cb:     cb,
		counts: &Counts{},
	}
	return c
}

func (s *closedState) onEnter() {
	cb := s.cb

	if cb.interval > 0 {
		s.ticker = time.NewTicker(cb.interval)
		s.stateLeaved = make(chan interface{}, 0)

		go func() {
			for {
				select {
				case <-s.ticker.C:
					if cb.currentState.getType() == Closed {
						s.counts.clear()
						cb.increaseGeneration()
					}
				case <-s.stateLeaved:
					return
				}
			}
		}()
	}
}

func (s *closedState) onLeave() {
	s.counts.clear()
	if s.cb.interval > 0 {
		s.ticker.Stop()
		close(s.stateLeaved)
	}
}

func (s *closedState) getType() StateType {
	return Closed
}

func (s *closedState) execute(req func() (interface{}, error)) (interface{}, error) {
	defer func() {
		e := recover()
		if e != nil {
			s.counts.onFailure()
			panic(e)
		}
	}()

	s.counts.onRequest()
	res, err := req()

	if err != nil {
		s.counts.onFailure()
		if s.cb.readyToTrip(*s.counts) {
			s.cb.changeState(open)
		}
		return nil, err
	}

	s.counts.onSuccess()
	return res, nil
}

func (s *closedState) getCounts() Counts {
	return *s.counts
}
