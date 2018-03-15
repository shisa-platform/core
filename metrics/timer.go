package metrics

import (
	"sync/atomic"
	"time"
)

// Timer measures time intervals.
type Timer struct {
	running uint32
	start   time.Time
	stop    time.Time
}

// Start begins the timer and clears any previous values.
func (t *Timer) Start() {
	if atomic.CompareAndSwapUint32(&t.running, 0, 1) {
		t.start = time.Now().UTC()
		t.stop = time.Time{}
	}
}

// Stop ends the timer and returns the elaspsed interval.
// Calling Stop on a Timer that has never run will always return
// zero, and calling Stop on a completed timer will return the
// previous value.
func (t *Timer) Stop() time.Duration {
	if atomic.CompareAndSwapUint32(&t.running, 1, 0) {
		t.stop = time.Now().UTC()

		return t.stop.Sub(t.start)
	}

	if t.start.IsZero() {
		return 0
	}

	return t.stop.Sub(t.start)
}

// Running returns true if this timer is currently running.
func (t *Timer) Running() bool {
	return atomic.LoadUint32(&t.running) == 1
}

// Reset will clear the timer if it is not running.  Calling
// Reset on a running timer does nothing.
func (t *Timer) Reset() {
	if atomic.LoadUint32(&t.running) == 1 {
		return
	}

	t.start = time.Time{}
	t.stop = time.Time{}
}

// Interval returns the elapsed time interval.
// Calling Interval on a Timer that has never run will always
// return zero.  Calling Interval on a running timer will return
// the interval from the start time to now.  Calling Interval on
// a stopped timer will always return the interval of the last
// run.
func (t *Timer) Interval() time.Duration {
	if t.start.IsZero() {
		return 0
	}

	if !t.stop.IsZero() {
		return t.stop.Sub(t.start)
	}

	return time.Now().UTC().Sub(t.start)
}

// Time runs the given function and returns the interval required
// to run.
func Time(f func()) time.Duration {
	var timer Timer
	timer.Start()
	f()
	return timer.Stop()
}
