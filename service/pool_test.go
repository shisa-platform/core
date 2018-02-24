package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	parent := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	cut := GetRequest(parent)
	defer PutRequest(cut)

	assert.NotNil(t, cut)
	assert.Equal(t, parent, cut.Request)
	assert.Nil(t, cut.PathParams)
	assert.Nil(t, cut.QueryParams)
	assert.Equal(t, "", cut.id)
	assert.Equal(t, "", cut.clientIP)
}
