package httpx

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/errorx"
)

// WriteResponse serializes a response instance to the
// ResponseWriter.
// `ResponseWriter.WriteHeader` will always be called with the
// value of `Response.StatusCode()` so it is not safe to use the
// `ResponseWriter` after calling this function.
// Any error returned from `Response.Serialize` will be returned.
func WriteResponse(w http.ResponseWriter, response Response) (err merry.Error) {
	defer errorx.CapturePanic(&err, "panic in response serializer")

	for k, vs := range response.Headers() {
		w.Header()[k] = vs
	}
	for k := range response.Trailers() {
		w.Header().Add("Trailer", k)
	}

	w.WriteHeader(response.StatusCode())

	err = response.Serialize(w)

	for k, vs := range response.Trailers() {
		w.Header()[k] = vs
	}

	return
}
