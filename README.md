gobreaker
=========
Implements the [Circuit Breaker pattern](https://msdn.microsoft.com/en-us/library/dn589784.aspx) in Go. This project is the rewrite of the 
[gobreaker](https://github.com/sony/gobreaker) using [state pattern](https://en.wikipedia.org/wiki/State_pattern).

### Key Changes in this edition
* Using state design pattern
* Using `timer.Ticker` for time-based expiration

### Not implemented yet
* [ ] `TwoStepCircuitBreaker`
