package httpx

import (
	"net/http"
	"time"

	"github.com/percolate/shisa/service"
)

// ResponseInterceptor implements `http.ResponseWriter` to capture
// the outgoing response status code and size.
type ResponseInterceptor struct {
	http.ResponseWriter
	start  time.Time
	status int
	size   int
}

func (i *ResponseInterceptor) Header() http.Header {
	return i.ResponseWriter.Header()
}

func (i *ResponseInterceptor) Write(data []byte) (int, error) {
	size, err := i.ResponseWriter.Write(data)
	i.size += size

	return size, err
}

func (i *ResponseInterceptor) WriteHeader(status int) {
	i.status = status
	i.ResponseWriter.WriteHeader(status)
}

// WriteResponse serializes a response instance
// The `WriteHeader` method will *always* be called with the
// value of `Response.StatusCode()` so it is not safe to use the
// `ResponseWriter` methods of this instance after calling this
// method.
// Any error returned from `Response.Serialize` will be returned.
func (i *ResponseInterceptor) WriteResponse(response service.Response) error {
	return WriteResponse(i, response)
}

// Flush attempts to call the `Flush` method on the underlying
// `ResponseWriter`, if it implments the `http.Flusher` interface.
func (i *ResponseInterceptor) Flush() ResponseSnapshot {
	if f, ok := i.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}

	return ResponseSnapshot{
		StatusCode: i.status,
		Size:       i.size,
		Start:      i.start,
		Elapsed:    time.Now().UTC().Sub(i.start),
		Metrics:    make(map[string]time.Duration),
	}
}

func NewInterceptor(w http.ResponseWriter) *ResponseInterceptor {
	return &ResponseInterceptor{
		ResponseWriter: w,
		status:         http.StatusOK,
		start:          time.Now().UTC(),
	}
}
