package httpx

import (
	"net/http"

	"github.com/percolate/shisa/service"
)

// WriteResponse serializes a response instance to the
// ResponseWriter.
// `ResponseWriter.WriteHeader` will always be called with the
// value of `Response.StatusCode()` so it is not safe to use the
// `ResponseWriter` after calling this function.
// Any error returned from `Response.Serialize` will be returned.
func WriteResponse(w http.ResponseWriter, response service.Response) (err error) {
	for k, vs := range response.Headers() {
		w.Header()[k] = vs
	}
	for k := range response.Trailers() {
		w.Header().Add("Trailer", k)
	}

	w.WriteHeader(response.StatusCode())

	_, err = response.Serialize(w)

	for k, vs := range response.Trailers() {
		w.Header()[k] = vs
	}

	return
}
