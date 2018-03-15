package metrics

import (
	"sync"
)

var (
	timerPool = sync.Pool{
		New: func() interface{} {
			return Timer{}
		},
	}
)

// Get returns a Timer instance from the shared pool, ready for
// (re)use.
func GetTimer() Timer {
	timer := timerPool.Get().(Timer)
	timer.Reset()

	return timer
}

// PutTimer returns the given tiemr back to the shared pool.
func PutTimer(timer Timer) {
	if timer.Running() {
		timer.Stop()
	}
	timerPool.Put(timer)
}
