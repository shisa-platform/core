package httpx

import (
	"bytes"
	"io"
	"net/http"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
)

// Handler is a block of logic to apply to a request.
// Returning a non-nil Response indicates request processing
// should stop.
type Handler func(context.Context, *Request) Response

func (h Handler) InvokeSafely(ctx context.Context, request *Request) (_ Response, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in handler")
	return h(ctx, request), nil
}

// ErrorHandler creates a response for the given error condition.
type ErrorHandler func(context.Context, *Request, merry.Error) Response

func (h ErrorHandler) InvokeSafely(ctx context.Context, request *Request, err merry.Error) (_ Response, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in handler")
	return h(ctx, request, err), nil
}

type responseBuffer struct {
	code    int
	headers http.Header
	body    *bytes.Buffer
	written bool
}

func (r *responseBuffer) StatusCode() int {
	return r.code
}

func (r *responseBuffer) Headers() http.Header {
	return r.headers
}

func (r *responseBuffer) Trailers() http.Header {
	return nil
}

func (r *responseBuffer) Err() error {
	return nil
}

func (r *responseBuffer) Serialize(w io.Writer) merry.Error {
	_, err := io.Copy(w, r.body)

	return merry.Wrap(err)
}

func (r *responseBuffer) Header() http.Header {
	return r.headers
}

func (r *responseBuffer) Write(bs []byte) (int, error) {
	r.written = true
	return r.body.Write(bs)
}

func (r *responseBuffer) WriteHeader(statusCode int) {
	r.written = true
	r.code = statusCode
}

// AdaptStandardHandler allows an http.Hander to be used as an
// httpx.Handler
func AdaptStandardHandler(handler http.Handler) Handler {
	return func(ctx context.Context, request *Request) Response {
		w := responseBuffer{
			code:    http.StatusOK,
			headers: make(http.Header),
			body:    new(bytes.Buffer),
		}
		r := request.Request.WithContext(ctx)

		handler.ServeHTTP(&w, r)
		if w.written {
			return &w
		}

		return nil
	}
}
