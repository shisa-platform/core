package auxillary

import (
	"io"
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

type unserializableResponse struct{}

func (r unserializableResponse) StatusCode() int {
	return 500
}

func (r unserializableResponse) Headers() http.Header {
	return nil
}

func (r unserializableResponse) Trailers() http.Header {
	return nil
}

func (r unserializableResponse) Serialize(io.Writer) (int, error) {
	return 0, merry.New("i blewed up")
}

type stubAuthorizer struct {
	ok  bool
	err merry.Error
}

func (a stubAuthorizer) Authorize(context.Context, *service.Request) (bool, merry.Error) {
	return a.ok, a.err
}
