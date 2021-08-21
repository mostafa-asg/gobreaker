package breaker

type halfOpenState struct {
	counts *Counts
	cb     *CircuitBreaker
}

func NewHalfOpenState(cb *CircuitBreaker) *halfOpenState {
	return &halfOpenState{
		cb:     cb,
		counts: &Counts{},
	}
}

func (s *halfOpenState) getType() StateType {
	return HalfOpen
}

func (s *halfOpenState) execute(req func() (interface{}, error)) (interface{}, error) {
	defer func() {
		e := recover()
		if e != nil {
			s.counts.onFailure()
			panic(e)
		}
	}()

	before := s.cb.generation
	s.counts.onRequest()
	res, err := req()
	if s.cb.generation != before {
		return res, err
	}

	if err != nil {
		s.counts.onFailure()
		s.cb.changeState(open)
		return nil, err
	}

	s.counts.onSuccess()
	cb := s.cb
	if s.counts.ConsecutiveSuccesses >= cb.maxRequests {
		cb.changeState(closed)
	}
	return res, nil
}

func (s *halfOpenState) getCounts() Counts {
	return *s.counts
}

func (s *halfOpenState) onEnter() {
	// nothing to do
}

func (s *halfOpenState) onLeave() {
	s.counts.clear()
}
