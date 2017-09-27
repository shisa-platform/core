package metrics

import (
	"testing"
)

func TestStopBeforeStart(t *testing.T) {
	var timer Timer
	duration := timer.Stop()
	if duration != 0 {
		t.Errorf("unexpected duration %d", duration)
	}
}

func TestStartStop(t *testing.T) {
	var timer Timer
	timer.Start()
	duration := timer.Stop()
	if duration <= 0 {
		t.Errorf("unexpected duration: %d", duration)
	}
}

func TestMultipleStart(t *testing.T) {
	var timer Timer
	timer.Start()
	timer.Start()
	duration := timer.Stop()
	if duration <= 0 {
		t.Errorf("unexpected duration: %d", duration)
	}
}

func TestMultipleStop(t *testing.T) {
	var timer Timer
	timer.Start()
	duration1 := timer.Stop()
	if duration1 <= 0 {
		t.Errorf("unexpected duration: %d", duration1)
	}
	duration2 := timer.Stop()
	if duration1 != duration2 {
		t.Errorf("unexpected duration: %d", duration2)
	}
}

func TestTimeClosure(t *testing.T) {
	var x uint64
	duration := Time(func() {
		x += 32
	})
	if duration <= 0 {
		t.Errorf("unexpected duration: %d", duration)
	}
	if x != 32 {
		t.Errorf("unexpected result: %d", x)
	}
}
