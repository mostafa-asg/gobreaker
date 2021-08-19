package breaker

import (
	"time"
)

const defaultInterval time.Duration = 0
const defaultTimeout = 60 * time.Second

func defaultReadyToTrip(counts Counts) bool {
	return counts.ConsecutiveFailures > 5
}

func defaultIsSuccessful(err error) bool {
	return err == nil
}
