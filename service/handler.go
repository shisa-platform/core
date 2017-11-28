package service

import (
	"github.com/percolate/shisa/context"
)

// xxx - fake proxy
// w.Header().Set("Content-Type", "text/plain")
// w.Write([]byte("hello, world"))

//go:generate charlatan -output=./handler_charlatan.go Handler

type Handler interface {
	Handle(context.Context, *Request) Response
}
