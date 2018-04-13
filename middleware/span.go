package middleware

import (
	"github.com/opentracing/opentracing-go"
)

var (
	noopSpan opentracing.Span
)

func init() {
	tracer := opentracing.NoopTracer{}
	noopSpan = tracer.StartSpan("noop")
}
