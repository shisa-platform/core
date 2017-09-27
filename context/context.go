package context

import (
	"context"
	"time"

	"github.com/percolate/shisa/log"
	"github.com/percolate/shisa/models"
)

const (
	IDKey     = "ContextRequestIDKey"
	ActorKey  = "ContextActorKey"
	LoggerKey = "ContextLoggerKey"
)

type IContext interface {
	context.Context
}

type ILoggableContext interface {
	IContext
	Info(string)
	Infof(string, ...interface{})
	Error(string)
	Errorf(string, ...interface{})
	Trace(string)
	Tracef(string, ...interface{})
}

type Context struct {
	context.Context
	RequestID string
	Actor     models.User
	Logger    logx.Logger
}

var _ IContext = &Context{}

func New(parent IContext, id string, actor models.User, log logx.Logger) *Context {
	return &Context{
		Context:   parent,
		RequestID: id,
		Actor:     actor,
		Logger:    log,
	}
}

func (c *Context) Info(message string) {
	c.Logger.Info(c.RequestID, message)
}

func (c *Context) Infof(format string, args ...interface{}) {
	c.Logger.Infof(c.RequestID, format, args...)
}

func (c *Context) Error(message string) {
	c.Logger.Error(c.RequestID, message)
}

func (c *Context) Errorf(format string, args ...interface{}) {
	c.Logger.Errorf(c.RequestID, format, args...)
}

func (c *Context) Trace(message string) {
	c.Logger.Trace(c.RequestID, message)
}

func (c *Context) Tracef(format string, args ...interface{}) {
	c.Logger.Tracef(c.RequestID, format, args...)
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
		case LoggerKey:
			return c.Logger
		}
	}

	return c.Context.Value(key)
}
