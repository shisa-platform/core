package httpx

import (
	"bytes"
	"net/http"
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
