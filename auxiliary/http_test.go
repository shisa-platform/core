package auxiliary

import (
	"io"
	"net/http"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

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

type mockErrorHandler struct {
	calls int
}

func (m *mockErrorHandler) Handle(context.Context, *service.Request, merry.Error) {
	m.calls++
}

func (m *mockErrorHandler) assertNotCalled(t *testing.T) {
	t.Helper()
	assert.Equal(t, 0, m.calls, "unexpected error handler calls")
}

func (m *mockErrorHandler) assertCalled(t *testing.T) {
	t.Helper()
	assert.NotEqual(t, 0, m.calls, "error handler not called")
}

func (m *mockErrorHandler) assertCalledN(t *testing.T, expected int) {
	t.Helper()
	assert.Equalf(t, expected, m.calls, "error handler called %d times, expected %d", m.calls, expected)
}
