package errorx

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
)

func TestPanicError(t *testing.T) {
	cut := Panic(merry.New("i blewed up!"), "uh-oh")
	assert.NotNil(t, cut)
	assert.Equal(t, "uh-oh: i blewed up!", cut.Error())
	assert.True(t, IsPanic(cut))
}

func TestPanicNonError(t *testing.T) {
	cut := Panic("somebody set us up the bomb!", "uh-oh")
	assert.NotNil(t, cut)
	assert.Equal(t, "uh-oh: \"somebody set us up the bomb!\"", cut.Error())
	assert.True(t, IsPanic(cut))
}

func TestPanicf(t *testing.T) {
	cut := Panicf("uh-oh", "zalgo: %q", "he comes")
	assert.NotNil(t, cut)
	assert.Equal(t, "zalgo: \"he comes\": \"uh-oh\"", cut.Error())
	assert.True(t, IsPanic(cut))
}

func TestIsPanic(t *testing.T) {
	assert.False(t, IsPanic(nil))
	assert.False(t, IsPanic(merry.New("i blewed up!")))
	assert.False(t, IsPanic(merry.New("i blewed up!").WithValue(sentinel, "fake")))
	assert.False(t, IsPanic(merry.New("i blewed up!").WithValue(sentinel, false)))

	assert.True(t, IsPanic(Panic(merry.New("i blewed up!"), "uh-oh")))
}
