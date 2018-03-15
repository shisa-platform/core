package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimerStopEmpty(t *testing.T) {
	var timer Timer

	assert.False(t, timer.Running())

	duration := timer.Stop()
	assert.Zero(t, duration)
}

func TestTimerStartStop(t *testing.T) {
	var timer Timer

	timer.Start()
	assert.True(t, timer.Running())
	assert.False(t, timer.start.IsZero())

	nap := time.Millisecond * 250
	time.Sleep(nap)

	duration := timer.Stop()
	assert.True(t, duration > nap)
	assert.False(t, timer.Running())
	assert.False(t, timer.stop.IsZero())
}

func TestTimerMultipleStart(t *testing.T) {
	var timer Timer

	timer.Start()
	start := timer.start

	timer.Start()
	assert.Equal(t, start, timer.start)

	duration := timer.Stop()
	assert.True(t, duration > 0)
}

func TestTimerMultipleStop(t *testing.T) {
	var timer Timer

	timer.Start()
	duration1 := timer.Stop()
	assert.True(t, duration1 > 0)

	duration2 := timer.Stop()
	assert.Equal(t, duration1, duration2)
}

func TestTimerInterval(t *testing.T) {
	var timer Timer

	timer.Start()
	lap := timer.Interval()
	assert.True(t, lap > 0)

	duration := timer.Stop()
	assert.True(t, duration > 0)
	assert.True(t, duration > lap)

	interval := timer.Interval()
	assert.Equal(t, duration, interval)
}

func TestTimerResetEmpty(t *testing.T) {
	var timer Timer

	timer.Reset()
	assert.True(t, timer.start.IsZero())
	assert.True(t, timer.stop.IsZero())
}

func TestTimerResetRunning(t *testing.T) {
	var timer Timer

	timer.Start()

	timer.Reset()
	assert.False(t, timer.start.IsZero())
	assert.True(t, timer.stop.IsZero())

	timer.Stop()
	assert.False(t, timer.stop.IsZero())
}

func TestTimerResetStoped(t *testing.T) {
	var timer Timer

	timer.Start()
	assert.False(t, timer.start.IsZero())

	timer.Stop()
	assert.False(t, timer.stop.IsZero())

	timer.Reset()
	assert.True(t, timer.start.IsZero())
	assert.True(t, timer.stop.IsZero())
}

func TestTimerIntervalEmpty(t *testing.T) {
	var timer Timer

	interval := timer.Interval()
	assert.Zero(t, interval)
}

func TestTimeClosure(t *testing.T) {
	var x uint64
	duration := Time(func() {
		x += 32
	})
	assert.True(t, duration > 0)
	assert.Equal(t, uint64(32), x)
}
