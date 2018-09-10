package context

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/shisa-platform/core/models"
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
	// RequestID returns the configured request id or an empty string
	RequestID() string
	// Actor returns the configured actor or nil
	Actor() models.User
	// Span returns the configured OpenTracing span or nil
	Span() opentracing.Span
	// StartSpan creates a new span with the given operation name and options.  The new span is not added to the current instance.
	// If the context instance contains a span it will be the parent of the new span and it's tracer will be used to create the span.  If no span is configured the global tracer will be used to create a new span.
	StartSpan(string, ...opentracing.StartSpanOption) opentracing.Span
	// WithSpan returns a copy of the current instance configured with the given parent
	WithParent(context.Context) Context
	// WithActor returns a copy of the current instance configured with the given actor
	WithActor(models.User) Context
	/// WithRequestID returns an copy of the current instance configured with the given request id
	WithRequestID(string) Context
	// WithSpan returns a copy of the current instance configured with the given span
	WithSpan(opentracing.Span) Context
	// WithValue returns a copy of the current instance configured with the given value
	WithValue(key, value interface{}) Context
	// WithCancel returns a copy of current instance with a new Done channel.
	WithCancel() (Context, context.CancelFunc)
	// WithDeadline returns a copy of the current instance with the deadline adjusted to be no later than the give time
	WithDeadline(time.Time) (Context, context.CancelFunc)
	// WithTimeout returns a copy of the current instance with the deadline adjusted to time.Now().Add(timeout)
	WithTimeout(time.Duration) (Context, context.CancelFunc)
}

type shisaCtx struct {
	context.Context
	requestID string
	actor     models.User
	span      opentracing.Span
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
	case SpanKey:
		if ctx.span != nil {
			return ctx.span
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
	case SpanKey:
		ctx.span = value.(opentracing.Span)
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

// New returns a new instance with the given parent
func New(parent context.Context) Context {
	return newShisaCtx(parent)
}

func newShisaCtx(parent context.Context) *shisaCtx {
	return &shisaCtx{
		Context: parent,
	}
}

// WithActor returns a new instance with the given parent and actor
func WithActor(parent context.Context, value models.User) Context {
	ctx := newShisaCtx(parent)
	ctx.actor = value
	return ctx
}

// WithRequestID returns a new instance with the given parent and request id
func WithRequestID(parent context.Context, value string) Context {
	ctx := newShisaCtx(parent)
	ctx.requestID = value
	return ctx
}

// WithSpan returns a new instance with the given parent and span
func WithSpan(parent context.Context, value opentracing.Span) Context {
	ctx := newShisaCtx(parent)
	ctx.span = value
	return ctx
}

// StartSpan returns a new instance with the given parent and a span created using the given operation name and options
func StartSpan(parent Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, Context) {
	ctx := newShisaCtx(parent)
	ctx.span = parent.StartSpan(operationName, opts...)

	return ctx.span, ctx
}

// WithValue returns a new instance with the given parent and value
func WithValue(parent context.Context, key, value interface{}) Context {
	ctx := newShisaCtx(parent)

	switch key {
	case IDKey:
		ctx.requestID = value.(string)
	case ActorKey:
		ctx.actor = value.(models.User)
	case SpanKey:
		ctx.span = value.(opentracing.Span)
	default:
		ctx.Context = context.WithValue(parent, key, value)
	}

	return ctx
}

// WithCancel returns a new instance and its cancel function with the given parent
func WithCancel(grandParent context.Context) (Context, context.CancelFunc) {
	parent, cancel := context.WithCancel(grandParent)
	ctx := newShisaCtx(parent)
	return ctx, cancel
}

// WithDeadline returns a new instance with the given deadline
func WithDeadline(grandParent context.Context, deadline time.Time) (Context, context.CancelFunc) {
	parent, cancel := context.WithDeadline(grandParent, deadline)
	ctx := newShisaCtx(parent)
	return ctx, cancel
}

// WithTimeout returns a new instance with the given timeout
func WithTimeout(grandParent context.Context, timeout time.Duration) (Context, context.CancelFunc) {
	parent, cancel := context.WithTimeout(grandParent, timeout)
	ctx := newShisaCtx(parent)
	return ctx, cancel
}
