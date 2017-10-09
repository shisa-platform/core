package context

import (
	"context"
	"time"

	"github.com/percolate/shisa/models"
)

const (
	IDKey    = "ContextRequestIDKey"
	ActorKey = "ContextActorKey"
)

//go:generate charlatan -interfaces=Context -output=./contexttest/charlatancontext.go /usr/local/go/src/context/context.go
type Context struct {
	context.Context
	RequestID string
	Actor     models.User
}

func New(parent context.Context, id string, actor models.User) *Context {
	return &Context{
		Context:   parent,
		RequestID: id,
		Actor:     actor,
	}
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *Context) Err() error {
	return c.Context.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	if name, ok := key.(string); ok {
		switch name {
		case IDKey:
			return c.RequestID
		case ActorKey:
			return c.Actor
		}
	}

	return c.Context.Value(key)
}
