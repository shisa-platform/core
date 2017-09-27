package handler

import (
	"ctx"
	"responses"
)

// xxx - fake proxy
// w.Header().Set("Content-Type", "text/plain")
// w.Write([]byte("hello, world"))

type Rules struct {
}

type Handler interface {
	Handle(ctx.Context, *http.Request) responses.Responder
}
