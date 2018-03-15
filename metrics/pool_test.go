package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	cut := GetTimer()
	defer PutTimer(cut)

	assert.False(t, cut.Running())
	assert.True(t, cut.start.IsZero())
	assert.True(t, cut.stop.IsZero())
}

func TestPoolReuse(t *testing.T) {
	cut := GetTimer()
	cut.Start()
	cut.Stop()
	PutTimer(cut)

	cut = GetTimer()
	assert.False(t, cut.Running())
	assert.True(t, cut.start.IsZero())
	assert.True(t, cut.stop.IsZero())
}

func TestPoolPut(t *testing.T) {
	cut := GetTimer()
	cut.Start()
	PutTimer(cut)

	assert.False(t, cut.Running())
}
