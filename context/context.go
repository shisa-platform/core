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

//go:generate charlatan -output=./context_charlatan.go Context

type Context interface {
	context.Context
	RequestID() string
	Actor() models.User
}

type ctx struct {
	context.Context
	requestID string
	actor     models.User
}

func New(parent context.Context, id string, actor models.User) Context {
	return &ctx{
		Context:   parent,
		requestID: id,
		actor:     actor,
	}
}

func (c *ctx) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *ctx) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *ctx) Err() error {
	return c.Context.Err()
}

func (c *ctx) Value(key interface{}) interface{} {
	if name, ok := key.(string); ok {
		switch name {
		case IDKey:
			return c.requestID
		case ActorKey:
			return c.actor
		}
	}

	return c.Context.Value(key)
}

func (c *ctx) RequestID() string {
	return c.requestID
}

func (c *ctx) Actor() models.User {
	return c.actor
}
