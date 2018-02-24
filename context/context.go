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
	WithParent(value context.Context) Context
	WithActor(value models.User) Context
	WithRequestID(value string) Context
	WithValue(key, value interface{}) Context
	WithDeadline(deadline time.Time) (Context, context.CancelFunc)
	WithTimeout(timeout time.Duration) (Context, context.CancelFunc)
}

type ctx struct {
	context.Context
	requestID string
	actor     models.User
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

func (c *ctx) RequestID() string {
	return c.requestID
}

func (c *ctx) Actor() models.User {
	return c.actor
}

func (c *ctx) Value(key interface{}) interface{} {
	switch key {
	case IDKey:
		return c.requestID
	case ActorKey:
		return c.actor
	}

	return c.Context.Value(key)
}

func (c *ctx) WithParent(value context.Context) Context {
	c.Context = value
	return c
}

func (c *ctx) WithActor(value models.User) Context {
	c.actor = value
	return c
}

func (c *ctx) WithRequestID(value string) Context {
	c.requestID = value
	return c
}

func (c *ctx) WithValue(key, value interface{}) Context {
	switch key {
	case IDKey:
		c.requestID = value.(string)
	case ActorKey:
		c.actor = value.(models.User)
	default:
		c.Context = context.WithValue(c.Context, key, value)
	}

	return c
}

func (c *ctx) WithDeadline(deadline time.Time) (Context, context.CancelFunc) {
	parent, cancel := context.WithDeadline(c.Context, deadline)
	c.Context = parent

	return c, cancel
}

func (c *ctx) WithTimeout(timeout time.Duration) (Context, context.CancelFunc) {
	parent, cancel := context.WithTimeout(c.Context, timeout)
	c.Context = parent

	return c, cancel
}

func New(parent context.Context) Context {
	return &ctx{Context: parent}
}

func WithActor(parent context.Context, value models.User) Context {
	return &ctx{Context: parent, actor: value}
}

func WithRequestID(parent context.Context, value string) Context {
	return &ctx{Context: parent, requestID: value}
}

func WithValue(parent context.Context, key, value interface{}) Context {
	c := &ctx{}
	switch key {
	case IDKey:
		c.requestID = value.(string)
	case ActorKey:
		c.actor = value.(models.User)
	default:
		parent = context.WithValue(parent, key, value)
	}

	c.Context = parent
	return c
}

func WithCancel(grandParent context.Context) (Context, context.CancelFunc) {
	parent, cancel := context.WithCancel(grandParent)
	c := &ctx{Context: parent}
	return c, cancel
}

func WithDeadline(grandParent context.Context, deadline time.Time) (Context, context.CancelFunc) {
	parent, cancel := context.WithDeadline(grandParent, deadline)
	c := &ctx{Context: parent}
	return c, cancel
}

func WithTimeout(grandParent context.Context, timeout time.Duration) (Context, context.CancelFunc) {
	parent, cancel := context.WithTimeout(grandParent, timeout)
	c := &ctx{Context: parent}
	return c, cancel
}
