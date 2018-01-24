package auxiliary

import (
	"io"
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func unserializableResponse() service.Response {
	return &service.FakeResponse{
		StatusCodeHook: func() int {
			return http.StatusInternalServerError
		},
		HeadersHook: func() http.Header {
			return nil
		},
		TrailersHook: func() http.Header {
			return nil
		},
		ErrHook: func() error {
			return nil
		},
		SerializeHook: func(io.Writer) (int, error) {
			return 0, merry.New("i blewed up!")
		},
	}
}

type stubAuthorizer struct {
	ok  bool
	err merry.Error
}

func (a stubAuthorizer) Authorize(context.Context, *service.Request) (bool, merry.Error) {
	return a.ok, a.err
}
