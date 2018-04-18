package errorx

import (
	"github.com/ansel1/merry"
)

type panicSentinel struct{}

var sentinel = new(panicSentinel)

/*
CapturePanic will capture a panic, create a merry.Error and set
exception to the error.  The error has message prepended to the value
originall passed to `panic`.

Use it like this:

    defer errorx.CapturePanic(&err, "panic while doing a thing")
*/
func CapturePanic(exception *merry.Error, message string) {
	arg := recover()
	if arg == nil {
		return
	}

	var ex merry.Error
	if err1, ok := arg.(error); ok {
		ex = merry.Prepend(err1, message)
	} else {
		ex = merry.Errorf("%s: %v", message, arg)
	}

	*exception = ex.WithValue(sentinel, true)

	return
}

// IsPanic returns true if the given error was created by CapturePanic
func IsPanic(err merry.Error) bool {
	value := merry.Value(err, sentinel)
	if value == nil {
		return false
	}

	if b, ok := value.(bool); ok && b {
		return true
	}

	return false
}
