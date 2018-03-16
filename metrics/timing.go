package metrics

import (
	"sync"
	"time"
)

type Timing struct {
	mux    sync.RWMutex
	timers map[string]*Timer
}

func NewTiming() *Timing {
	return &Timing{
		timers: make(map[string]*Timer),
	}
}

// Start begins the named timer (if it is not already running)
// and clears any previous values.  If the named timer doesn't
// exist it a new one be created and started.
func (t *Timing) Start(name string) {
	t.mux.RLock()
	if timer, ok := t.timers[name]; ok {
		timer.Start()
		t.mux.RUnlock()
		return
	}
	t.mux.RUnlock()

	t.mux.Lock()
	defer t.mux.Unlock()
	if timer, ok := t.timers[name]; ok {
		timer.Start()
		return
	}

	timer := &Timer{}
	timer.Start()
	t.timers[name] = timer
}

// Stop ends the named timer and returns the elaspsed interval.
// The result is zero if the named timer doesn't exist.
// Calling Stop on a completed timer will return the previous
// value.
func (t *Timing) Stop(name string) time.Duration {
	t.mux.RLock()
	defer t.mux.RUnlock()

	if timer, ok := t.timers[name]; ok {
		return timer.Stop()
	}

	return 0
}

// Running returns true if the name timer is currently running.
// The result is false if the named timer doesn't exit.
func (t *Timing) Running(name string) bool {
	t.mux.RLock()
	defer t.mux.RUnlock()

	if timer, ok := t.timers[name]; ok {
		return timer.Running()
	}

	return false
}

// Reset will clear the named timer if it is not running.
// Calling Reset on a running timer does nothing.
func (t *Timing) Reset(name string) {
	t.mux.RLock()
	defer t.mux.RUnlock()

	if timer, ok := t.timers[name]; ok {
		timer.Reset()
	}
}

// ResetAll will clear all the timers.
func (t *Timing) ResetAll() {
	t.mux.RLock()
	defer t.mux.RUnlock()

	for _, timer := range t.timers {
		timer.Reset()
	}
}

// Do calls f for each timer.
func (t *Timing) Do(f func(string, *Timer)) {
	t.mux.RLock()
	defer t.mux.RUnlock()

	for name, timer := range t.timers {
		f(name, timer)
	}
}

// Interval returns the elapsed time interval for the named
// timer. The result is zero if the named timer doesn't exist.
// Calling Interval on a running timer will return the interval
// from the start time to now.  Calling Interval on a stopped
// timer will always return the interval of the last run.
func (t *Timing) Interval(name string) time.Duration {
	t.mux.RLock()
	defer t.mux.RUnlock()

	if timer, ok := t.timers[name]; ok {
		return timer.Interval()
	}

	return 0
}

// Delete removes the named timer.
func (t *Timing) Delete(name string) {
	t.mux.Lock()
	defer t.mux.Unlock()

	delete(t.timers, name)
}
