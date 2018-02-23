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

func unserializableResponse() httpx.Response {
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

func (a stubAuthorizer) Authorize(context.Context, *httpx.Request) (bool, merry.Error) {
	return a.ok, a.err
}

type mockErrorHook struct {
	calls int
}

func (m *mockErrorHook) Handle(context.Context, *httpx.Request, merry.Error) {
	m.calls++
}

func (m *mockErrorHook) assertNotCalled(t *testing.T) {
	t.Helper()
	assert.Equal(t, 0, m.calls, "unexpected error handler calls")
}

func (m *mockErrorHook) assertCalled(t *testing.T) {
	t.Helper()
	assert.NotEqual(t, 0, m.calls, "error handler not called")
}

func (m *mockErrorHook) assertCalledN(t *testing.T, expected int) {
	t.Helper()
	assert.Equalf(t, expected, m.calls, "error handler called %d times, expected %d", m.calls, expected)
}
