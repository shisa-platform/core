package context

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/percolate/shisa/models"
)

type idKey struct{}
type actorKey struct{}
type spanKey struct{}

var (
	IDKey    = new(idKey)
	ActorKey = new(actorKey)
	SpanKey  = new(spanKey)
)

//go:generate charlatan -output=./context_charlatan.go Context

type Context interface {
	context.Context
	RequestID() string
	Actor() models.User
	Span() opentracing.Span
	StartSpan(string, ...opentracing.StartSpanOption) opentracing.Span
	WithParent(context.Context) Context
	WithActor(models.User) Context
	WithRequestID(string) Context
	WithSpan(opentracing.Span) Context
	WithValue(key, value interface{}) Context
	WithCancel() (Context, context.CancelFunc)
	WithDeadline(time.Time) (Context, context.CancelFunc)
	WithTimeout(time.Duration) (Context, context.CancelFunc)
}

type shisaCtx struct {
	context.Context
	requestID string
	actor     models.User
	span      opentracing.Span
}

func (ctx *shisaCtx) Deadline() (deadline time.Time, ok bool) {
	return ctx.Context.Deadline()
}

func (ctx *shisaCtx) Done() <-chan struct{} {
	return ctx.Context.Done()
}

func (ctx *shisaCtx) Err() error {
	return ctx.Context.Err()
}

func (ctx *shisaCtx) RequestID() string {
	if ctx.requestID != "" {
		return ctx.requestID
	}

	if value := ctx.Context.Value(IDKey); value != nil {
		return value.(string)
	}

	return ""
}

func (ctx *shisaCtx) Actor() models.User {
	if ctx.actor != nil {
		return ctx.actor
	}

	if value := ctx.Context.Value(ActorKey); value != nil {
		return value.(models.User)
	}

	return nil
}

func (ctx *shisaCtx) Span() opentracing.Span {
	if ctx.span != nil {
		return ctx.span
	}

	if value := ctx.Context.Value(SpanKey); value != nil {
		return value.(opentracing.Span)
	}

	return opentracing.SpanFromContext(ctx.Context)
}

func (ctx *shisaCtx) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	if ctx.span != nil {
		opts = append(opts, opentracing.ChildOf(ctx.span.Context()))
		tracer = ctx.span.Tracer()
	}

	return tracer.StartSpan(operationName, opts...)
}

func (ctx *shisaCtx) Value(key interface{}) interface{} {
	switch key {
	case IDKey:
		if ctx.requestID != "" {
			return ctx.requestID
		}
	case ActorKey:
		if ctx.actor != nil {
			return ctx.actor
		}
	}

	return ctx.Context.Value(key)
}

func (ctx *shisaCtx) WithParent(value context.Context) Context {
	ctx.Context = value
	return ctx
}

func (ctx *shisaCtx) WithActor(value models.User) Context {
	ctx.actor = value
	return ctx
}

func (ctx *shisaCtx) WithRequestID(value string) Context {
	ctx.requestID = value
	return ctx
}

func (ctx *shisaCtx) WithSpan(value opentracing.Span) Context {
	ctx.span = value
	return ctx
}

func (ctx *shisaCtx) WithValue(key, value interface{}) Context {
	switch key {
	case IDKey:
		ctx.requestID = value.(string)
	case ActorKey:
		ctx.actor = value.(models.User)
	default:
		ctx.Context = context.WithValue(ctx.Context, key, value)
	}

	return ctx
}

func (ctx *shisaCtx) WithCancel() (Context, context.CancelFunc) {
	parent, cancel := context.WithCancel(ctx.Context)
	ctx.Context = parent

	return ctx, cancel
}

func (ctx *shisaCtx) WithDeadline(deadline time.Time) (Context, context.CancelFunc) {
	parent, cancel := context.WithDeadline(ctx.Context, deadline)
	ctx.Context = parent

	return ctx, cancel
}

func (ctx *shisaCtx) WithTimeout(timeout time.Duration) (Context, context.CancelFunc) {
	parent, cancel := context.WithTimeout(ctx.Context, timeout)
	ctx.Context = parent

	return ctx, cancel
}

func New(parent context.Context) Context {
	return get(parent)
}

func WithActor(parent context.Context, value models.User) Context {
	ctx := get(parent)
	ctx.actor = value
	return ctx
}

func WithRequestID(parent context.Context, value string) Context {
	ctx := get(parent)
	ctx.requestID = value
	return ctx
}

func WithSpan(parent context.Context, value opentracing.Span) Context {
	ctx := get(parent)
	ctx.span = value
	return ctx
}

func StartSpan(parent Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, Context) {
	ctx := get(parent)
	ctx.span = parent.StartSpan(operationName, opts...)

	return ctx.span, ctx
}

func WithValue(parent context.Context, key, value interface{}) Context {
	ctx := get(parent)

	switch key {
	case IDKey:
		ctx.requestID = value.(string)
	case ActorKey:
		ctx.actor = value.(models.User)
	default:
		ctx.Context = context.WithValue(parent, key, value)
	}

	return ctx
}

func WithCancel(grandParent context.Context) (Context, context.CancelFunc) {
	parent, cancel := context.WithCancel(grandParent)
	ctx := get(parent)
	return ctx, cancel
}

func WithDeadline(grandParent context.Context, deadline time.Time) (Context, context.CancelFunc) {
	parent, cancel := context.WithDeadline(grandParent, deadline)
	ctx := get(parent)
	return ctx, cancel
}

func WithTimeout(grandParent context.Context, timeout time.Duration) (Context, context.CancelFunc) {
	parent, cancel := context.WithTimeout(grandParent, timeout)
	ctx := get(parent)
	return ctx, cancel
}
