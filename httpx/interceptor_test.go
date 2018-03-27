package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
)

func TestInterceptor(t *testing.T) {
	rw := httptest.NewRecorder()
	cut := NewInterceptor(rw)

	cut.Header().Set("x-zalgo", "he comes")
	assert.Equal(t, "he comes", rw.HeaderMap.Get("X-Zalgo"))

	cut.WriteHeader(451)
	assert.Equal(t, 451, rw.Code)

	body := "the <center> cannot hold"
	cut.Write([]byte(body))
	assert.Equal(t, body, rw.Body.String())

	assert.True(t, cut.Elapsed() > 0)

	snapshot := cut.Flush()
	assert.True(t, rw.Flushed)
	assert.Equal(t, 451, snapshot.StatusCode)
	assert.Equal(t, len(body), snapshot.Size)
	assert.True(t, snapshot.Elapsed > 0)

	dupeshot := cut.Snapshot()
	assert.Equal(t, snapshot.StatusCode, dupeshot.StatusCode)
	assert.Equal(t, snapshot.Size, dupeshot.Size)
	assert.NotEqual(t, snapshot.Elapsed, dupeshot.Elapsed)
	assert.Equal(t, snapshot.Start, dupeshot.Start)
}

func TestInterceptorWriteResponse(t *testing.T) {
	rw := httptest.NewRecorder()
	cut := NewInterceptor(rw)

	response := NewEmpty(451)
	response.Headers().Set("x-zalgo", "he comes")
	response.Trailers().Set("x-signal", "main screen turn on")

	assert.NoError(t, cut.WriteResponse(response))
	assert.Equal(t, 451, rw.Code)
	assert.Zero(t, rw.Body.Len())
	assert.Equal(t, "he comes", rw.HeaderMap.Get("X-Zalgo"))
	assert.Equal(t, "main screen turn on", rw.HeaderMap.Get("X-Signal"))
}

func TestInterceptorWriteResponsePanic(t *testing.T) {
	rw := httptest.NewRecorder()
	cut := NewInterceptor(rw)

	response := &FakeResponse{
		StatusCodeHook: func() int { return 451 },
		HeadersHook:    func() http.Header { return nil },
		TrailersHook:   func() http.Header { return nil },
		ErrHook:        func() error { return nil },
		SerializeHook: func(io.Writer) merry.Error {
			panic(merry.New("i blewed up"))
		},
	}

	assert.Error(t, cut.WriteResponse(response))
}

func TestInterceptorWriteResponsePanicString(t *testing.T) {
	rw := httptest.NewRecorder()
	cut := NewInterceptor(rw)

	response := &FakeResponse{
		StatusCodeHook: func() int { return 451 },
		HeadersHook:    func() http.Header { return nil },
		TrailersHook:   func() http.Header { return nil },
		ErrHook:        func() error { return nil },
		SerializeHook:  func(io.Writer) merry.Error { panic("i blewed up") },
	}

	assert.Error(t, cut.WriteResponse(response))
}
