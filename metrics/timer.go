package metrics

import (
	"time"
)

// Timer measures time intervals in nanoseconds.
// The ticks of time measured are not subject to either jumps or slew by NTP adjustments made on the host system.  Therefore, only very short intervals (ideally sub-second) should be measured.  Due to inconsistencies in the accuracy of system oscillators long intervals will exhibit a mean divergence from the length of a SI second.
// A monotonic clock source is used internally which is relative to an arbitrary point, so the internal start and end timestamps *cannot* be converted to civil time.  This code will only run under Linux currently as it makes a syscall into the kernel to retrieve timestamps.  See this Github Issue for more background: https://github.com/golang/go/issues/12914
type Timer struct {
	start int64
	stop  int64
}

// Start begins the timer and clears any previous end time.
func (t *Timer) Start() {
	t.start = GetMonotonicRawTime()
	t.stop = 0
}

// Stop ends the timer and returns the elaspsed interval in nanoseconds.
// Calling Stop on a Timer that has never run will always return zero, and calling Stop on a completed timer will return the previous value.
func (t *Timer) Stop() time.Duration {
	if t.start == 0 {
		return 0
	}
	if t.stop != 0 {
		return time.Duration(t.stop - t.start)
	}
	t.stop = GetMonotonicRawTime()

	return time.Duration(t.stop - t.start)
}

// Time runs the given function and returns the interval required to run it.
func Time(f func()) time.Duration {
	var timer Timer
	timer.Start()
	f()
	return timer.Stop()
}
