package handler

import (
	"net/http"

	"github.com/percolate/shisa/context"
)

// xxx - fake proxy
// w.Header().Set("Content-Type", "text/plain")
// w.Write([]byte("hello, world"))

type Rules struct {
}

type Handler interface {
	Handle(context.Context, *http.Request) //responses.Responder
}
