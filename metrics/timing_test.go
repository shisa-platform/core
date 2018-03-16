package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimingStartStop(t *testing.T) {
	timing := NewTiming()

	timing.Start("test")
	timer, ok := timing.timers["test"]
	assert.NotNil(t, timer)
	assert.True(t, ok)
	assert.True(t, timer.Running())
	assert.True(t, timing.Running("test"))
	assert.False(t, timer.start.IsZero())

	nap := time.Millisecond * 250
	time.Sleep(nap)

	duration := timing.Stop("test")
	assert.True(t, duration > nap)
	assert.False(t, timing.Running("test"))
	assert.False(t, timer.stop.IsZero())
}

func TestTimingMultipleStart(t *testing.T) {
	timing := NewTiming()

	timing.Start("test")
	timer, ok := timing.timers["test"]
	assert.NotNil(t, timer)
	assert.True(t, ok)
	start := timer.start

	timing.Start("test")
	assert.Equal(t, start, timer.start)

	duration := timing.Stop("test")
	assert.True(t, duration > 0)
}

func TestTimingStopMissing(t *testing.T) {
	timing := NewTiming()

	assert.False(t, timing.Running("test"))

	duration := timing.Stop("test")
	assert.Zero(t, duration)
}

func TestTimingMultipleStop(t *testing.T) {
	timing := NewTiming()

	timing.Start("test")
	duration1 := timing.Stop("test")
	assert.True(t, duration1 > 0)

	duration2 := timing.Stop("test")
	assert.Equal(t, duration1, duration2)
}

func TestTimingIntervalMissing(t *testing.T) {
	timing := NewTiming()

	interval := timing.Interval("test")
	assert.Zero(t, interval)
}

func TestTimingInterval(t *testing.T) {
	timing := NewTiming()

	timing.Start("test")
	lap := timing.Interval("test")
	assert.True(t, lap > 0)

	duration := timing.Stop("test")
	assert.True(t, duration > 0)
	assert.True(t, duration > lap)

	interval := timing.Interval("test")
	assert.Equal(t, duration, interval)
}

func TestTimingResetMissing(t *testing.T) {
	timing := NewTiming()

	timing.Reset("test")
	assert.Len(t, timing.timers, 0)
}

func TestTimingResetRunning(t *testing.T) {
	timing := NewTiming()

	timing.Start("test")
	timer, ok := timing.timers["test"]
	assert.NotNil(t, timer)
	assert.True(t, ok)

	timing.Reset("test")
	assert.False(t, timer.start.IsZero())
	assert.True(t, timer.stop.IsZero())

	timing.Stop("test")
	assert.False(t, timer.stop.IsZero())
}

func TestTimingResetStoped(t *testing.T) {
	timing := NewTiming()

	timing.Start("test")
	timer, ok := timing.timers["test"]
	assert.NotNil(t, timer)
	assert.True(t, ok)
	assert.False(t, timer.start.IsZero())

	timing.Stop("test")
	assert.False(t, timer.stop.IsZero())

	timing.Reset("test")
	assert.True(t, timer.start.IsZero())
	assert.True(t, timer.stop.IsZero())
}

func TestTimingResetAll(t *testing.T) {
	timing := NewTiming()

	timing.Start("foo")
	timing.Stop("foo")
	timing.Start("bar")
	timing.Stop("bar")

	timing.ResetAll()
	assert.Zero(t, timing.Interval("foo"))
	assert.Zero(t, timing.Interval("bar"))
}

func TestTimingDeleteMissing(t *testing.T) {
	timing := NewTiming()

	timing.Start("foo")
	timing.Stop("foo")
	assert.Len(t, timing.timers, 1)

	timing.Delete("test")
	assert.Len(t, timing.timers, 1)
}

func TestTimingDelete(t *testing.T) {
	timing := NewTiming()

	timing.Start("test")
	timing.Stop("test")
	assert.Len(t, timing.timers, 1)

	timing.Delete("test")
	assert.Len(t, timing.timers, 0)
}

func TestTimingDo(t *testing.T) {
	timing := NewTiming()

	timing.Start("foo")
	timing.Stop("foo")
	timing.Start("bar")
	timing.Stop("bar")

	count := 0
	timing.Do(func(string, *Timer) {
		count++
	})
	assert.Equal(t, 2, count)
}
