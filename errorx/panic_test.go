package errorx

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
)

func TestPanicError(t *testing.T) {
	var exception merry.Error

	defer func() {
		assert.Error(t, exception)
		assert.Equal(t, "uh-oh: i blewed up!", exception.Error())
		assert.True(t, IsPanic(exception))
	}()

	defer CapturePanic(&exception, "uh-oh")

	panic(merry.New("i blewed up!"))
}

func TestPanicNonError(t *testing.T) {
	var exception merry.Error

	defer func() {
		assert.Error(t, exception)
		assert.Equal(t, "uh-oh: somebody set us up the bomb!", exception.Error())
		assert.True(t, IsPanic(exception))
	}()

	defer CapturePanic(&exception, "uh-oh")

	panic("somebody set us up the bomb!")
}

func TestPanicNoError(t *testing.T) {
	var exception merry.Error

	defer func() {
		assert.NoError(t, exception)
	}()

	defer CapturePanic(&exception, "uh-oh")
}

func TestIsPanic(t *testing.T) {
	assert.False(t, IsPanic(nil))
	assert.False(t, IsPanic(merry.New("i blewed up!")))
	assert.False(t, IsPanic(merry.New("i blewed up!").WithValue(sentinel, "fake")))
	assert.False(t, IsPanic(merry.New("i blewed up!").WithValue(sentinel, false)))

	var exception merry.Error

	defer func() {
		assert.True(t, IsPanic(exception))
	}()

	defer CapturePanic(&exception, "uh-oh")

	panic(merry.New("i blewed up!"))
}
