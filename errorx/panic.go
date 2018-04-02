package errorx

import (
	"fmt"

	"github.com/ansel1/merry"
)

const (
	sentinel = "panic"
)

func Panic(arg interface{}, message string) (err merry.Error) {
	if err1, ok := arg.(error); ok {
		err = merry.Prepend(err1, message)
	} else {
		err = merry.Errorf("%s: \"%v\"", message, arg)
	}

	err = err.WithValue(sentinel, true)

	return
}

func Panicf(arg interface{}, format string, args ...interface{}) merry.Error {
	return Panic(arg, fmt.Sprintf(format, args...))
}

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
