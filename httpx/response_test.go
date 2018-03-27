package httpx

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
)

func TestErrorResponse(t *testing.T) {
	err := merry.New("i blewed up")
	cut := NewEmptyError(http.StatusInternalServerError, err)

	assert.NotNil(t, cut)
	assert.Equal(t, http.StatusInternalServerError, cut.StatusCode())
	assert.Empty(t, cut.Headers())
	assert.Empty(t, cut.Trailers())
	assert.Error(t, cut.Err())
	assert.Equal(t, err, cut.Err())

	var buf bytes.Buffer
	serr := cut.Serialize(&buf)
	assert.NoError(t, serr)
	assert.Equal(t, 0, buf.Len())
}

func TestJsonResponse(t *testing.T) {
	body := `{"zalgo": "he comes"}`
	cut := NewOK(json.RawMessage(body))

	rw := httptest.NewRecorder()
	assert.NoError(t, WriteResponse(rw, cut))

	assert.JSONEq(t, body, rw.Body.String())
}

func TestSeeOtherResponse(t *testing.T) {
	location := "https://example.com"
	cut := NewSeeOther(location)

	assert.Equal(t, http.StatusSeeOther, cut.StatusCode())
	assert.Equal(t, location, cut.Headers().Get(LocationHeaderKey))

	rw := httptest.NewRecorder()
	assert.NoError(t, WriteResponse(rw, cut))

	assert.Equal(t, http.StatusSeeOther, rw.Code)
	assert.Equal(t, location, rw.HeaderMap.Get(LocationHeaderKey))
}

func TestTemporaryRedirectResponse(t *testing.T) {
	location := "https://example.com"
	cut := NewTemporaryRedirect(location)

	assert.Equal(t, http.StatusTemporaryRedirect, cut.StatusCode())
	assert.Equal(t, location, cut.Headers().Get(LocationHeaderKey))

	rw := httptest.NewRecorder()
	assert.NoError(t, WriteResponse(rw, cut))

	assert.Equal(t, http.StatusTemporaryRedirect, rw.Code)
	assert.Equal(t, location, rw.HeaderMap.Get(LocationHeaderKey))
}

func TestResponseAdapter(t *testing.T) {
	payload := `{"zalgo": "he comes"}`
	cut := &ResponseAdapter{
		Response: &http.Response{
			Status:        "451 Unavailable For Legal Reasons",
			StatusCode:    451,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Body:          ioutil.NopCloser(bytes.NewBufferString(payload)),
			ContentLength: int64(len(payload)),
			Uncompressed:  true,
		},
	}
	assert.Nil(t, cut.Err())

	rw := httptest.NewRecorder()
	assert.NoError(t, WriteResponse(rw, cut))

	assert.Equal(t, 451, rw.Code)
	assert.JSONEq(t, payload, rw.Body.String())
}
