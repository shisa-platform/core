package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
)

func TestRequestParseQueryParametersEmptyQuery(t *testing.T) {
	url := "http://example.com/test"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())

	assert.Len(t, cut.QueryParams, 0)
}

func TestRequestParseQueryParametersIgnoreEmptyQueryPair(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&&waits=behind%20the%20walls&zalgo=slithy"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())

	assert.Len(t, cut.QueryParams, 2)
	for i, p := range cut.QueryParams {
		switch i {
		case 0:
			assert.Len(t, p.Values, 2)
			assert.Equal(t, "zalgo", p.Name)
			assert.Equal(t, "he:comes", p.Values[0])
			assert.Equal(t, "slithy", p.Values[1])
			assert.NoError(t, p.Err)
		case 1:
			assert.Len(t, p.Values, 1)
			assert.Equal(t, "waits", p.Name)
			assert.Equal(t, "behind the walls", p.Values[0])
			assert.NoError(t, p.Err)
		}
	}
}

func TestRequestParseQueryParametersBadEscapeCodes(t *testing.T) {
	url := "http://example.com/test?name=foo%zzbar&sec%zzret=42"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.False(t, cut.ParseQueryParameters())

	assert.Len(t, cut.QueryParams, 2)
	for i, p := range cut.QueryParams {
		switch i {
		case 0:
			assert.Len(t, p.Values, 1)
			assert.Equal(t, "name", p.Name)
			assert.Equal(t, "foo%zzbar", p.Values[0])
			assert.Error(t, p.Err)
		case 1:
			assert.Len(t, p.Values, 1)
			assert.Equal(t, "sec%zzret", p.Name)
			assert.Equal(t, "42", p.Values[0])
			assert.Error(t, p.Err)
		}
	}
}

func TestRequestValidateQueryParametersPanic(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 2)

	validator := func([]string) merry.Error {
		panic(merry.New("i blewed up!"))
	}
	fields := []Field{{Name: "zalgo"}, {Name: "waits", Validator: validator}}

	_, _, err := cut.ValidateQueryParameters(fields)
	assert.Error(t, err)
	assert.NoError(t, cut.QueryParams[0].Err)
	assert.NoError(t, cut.QueryParams[1].Err)
}

func TestRequestValidateQueryParametersAlreadyBad(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 2)

	expectedErr := merry.New("something bad")
	cut.QueryParams[1].Err = expectedErr

	validator := func([]string) merry.Error {
		return merry.New("meh")
	}
	fields := []Field{{Name: "zalgo"}, {Name: "waits", Validator: validator}}

	malformed, unknown, err := cut.ValidateQueryParameters(fields)
	assert.True(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, err)
	assert.NoError(t, cut.QueryParams[0].Err)
	assert.Error(t, cut.QueryParams[1].Err)
	assert.Equal(t, expectedErr, cut.QueryParams[1].Err)
}

func TestRequestValidateQueryParametersInvalidParameter(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 2)

	validator := func([]string) merry.Error {
		return merry.New("meh")
	}
	fields := []Field{{Name: "zalgo"}, {Name: "waits", Validator: validator}}

	malformed, unknown, err := cut.ValidateQueryParameters(fields)
	assert.True(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, err)
	assert.NoError(t, cut.QueryParams[0].Err)
	assert.Error(t, cut.QueryParams[1].Err)
	assert.True(t, merry.Is(cut.QueryParams[1].Err, MalformedQueryParamter))
}

func TestRequestValidateQueryParametersUnknownParameter(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 2)

	fields := []Field{{Name: "zalgo"}}

	malformed, unknown, err := cut.ValidateQueryParameters(fields)
	assert.False(t, malformed)
	assert.True(t, unknown)
	assert.NoError(t, err)
	assert.NoError(t, cut.QueryParams[0].Err)
	assert.Error(t, cut.QueryParams[1].Err)
	assert.True(t, merry.Is(cut.QueryParams[1].Err, UnknownQueryParamter))
}

func TestRequestValidateQueryParametersInvalidAndUnknownParameters(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 2)

	validator := func([]string) merry.Error {
		return merry.New("meh")
	}
	fields := []Field{{Name: "zalgo", Validator: validator}}

	malformed, unknown, err := cut.ValidateQueryParameters(fields)
	assert.True(t, malformed)
	assert.True(t, unknown)
	assert.NoError(t, err)
	assert.Error(t, cut.QueryParams[0].Err)
	assert.True(t, merry.Is(cut.QueryParams[0].Err, MalformedQueryParamter))
	assert.Error(t, cut.QueryParams[1].Err)
	assert.True(t, merry.Is(cut.QueryParams[1].Err, UnknownQueryParamter))
}

func TestRequestValidateQueryParametersMissingParameterWithDefault(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 1)

	fields := []Field{{Name: "zalgo"}, {Name: "waits", Default: "behind the walls"}}

	malformed, unknown, err := cut.ValidateQueryParameters(fields)
	assert.False(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, err)
	assert.Len(t, cut.QueryParams, 2)
	assert.NoError(t, cut.QueryParams[0].Err)
	assert.NoError(t, cut.QueryParams[1].Err)
	assert.Equal(t, "waits", cut.QueryParams[1].Name)
	assert.ElementsMatch(t, []string{"behind the walls"}, cut.QueryParams[1].Values)
}

func TestRequestValidateQueryParametersMissingRequiredParameter(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 1)

	fields := []Field{{Name: "zalgo"}, {Name: "waits", Required: true}}

	malformed, unknown, err := cut.ValidateQueryParameters(fields)
	assert.True(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, err)
	assert.Len(t, cut.QueryParams, 2)
	assert.NoError(t, cut.QueryParams[0].Err)
	assert.Error(t, cut.QueryParams[1].Err)
	assert.True(t, merry.Is(cut.QueryParams[1].Err, MissingQueryParamter))
	assert.Equal(t, "waits", cut.QueryParams[1].Name)
}

func TestRequestValidateQueryParametersMissingParameter(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 1)

	fields := []Field{{Name: "zalgo"}, {Name: "waits"}}

	malformed, unknown, err := cut.ValidateQueryParameters(fields)
	assert.False(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, err)
	assert.Len(t, cut.QueryParams, 1)
	assert.NoError(t, cut.QueryParams[0].Err)
}

func TestRequestValidateQueryParameters(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 2)

	fields := []Field{{Name: "zalgo"}, {Name: "waits"}}

	malformed, unknown, err := cut.ValidateQueryParameters(fields)
	assert.False(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, err)
	assert.NoError(t, cut.QueryParams[0].Err)
	assert.NoError(t, cut.QueryParams[1].Err)
}
