package context

import (
	"context"
	"sync"
)

var (
	contextPool = sync.Pool{
		New: func() interface{} {
			return new(shisaCtx)
		},
	}
)

// Get returns a Context instance from the shared pool, ready for
// (re)use.
func Get(parent context.Context) Context {
	return get(parent)
}

func get(parent context.Context) *shisaCtx {
	ctx := contextPool.Get().(*shisaCtx)
	ctx.Context = parent
	ctx.requestID = ""
	ctx.actor = nil
	ctx.span = nil

	return ctx
}

// Put returns the given Context back to the shared pool.
func Put(ctx Context) {
	if c, ok := ctx.(*shisaCtx); ok {
		c.Context = nil
		contextPool.Put(c)
	}
}
