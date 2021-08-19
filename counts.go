package breaker

import "sync"

// Counts holds the numbers of requests and their successes/failures.
// CircuitBreaker clears the internal Counts either
// on the change of the state or at the closed-state intervals.
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
	mutex                sync.Mutex
}

func NewCounts(r, ts, tf, cs, cf uint32) Counts {
	return Counts{
		Requests:             r,
		TotalSuccesses:       ts,
		TotalFailures:        tf,
		ConsecutiveSuccesses: cs,
		ConsecutiveFailures:  cf,
	}
}

func (c *Counts) onRequest() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Requests++
}

func (c *Counts) onFailure() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

func (c *Counts) onSuccess() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

func (c *Counts) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}
