package context

import (
	"context"
	"sync"
)

var (
	contextPool = sync.Pool{
		New: func() interface{} {
			return new(ctx)
		},
	}
)

// Get returns a Context instance from the shared pool, ready for
// (re)use.
func Get(parent context.Context) Context {
	ctx := contextPool.Get().(*ctx)
	ctx.Context = parent
	ctx.requestID = ""
	ctx.actor = nil

	return ctx
}

// Put returns the given Context back to the shared pool.
func Put(ctx Context) {
	contextPool.Put(ctx)
}
