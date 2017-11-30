package service

import (
	"github.com/percolate/shisa/context"
)

//go:generate charlatan -output=./handler_charlatan.go Handler

type Handler interface {
	Handle(context.Context, *Request) Response
}
