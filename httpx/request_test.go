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

func TestRequestParseQueryParametersMultiplicity(t *testing.T) {
	url := "http://example.com/test?name=me&name=you&he=too"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())

	assert.Len(t, cut.QueryParams, 2)
	for i, p := range cut.QueryParams {
		switch i {
		case 0:
			assert.Len(t, p.Values, 2)
			assert.Equal(t, "name", p.Name)
			assert.Equal(t, "me", p.Values[0])
			assert.Equal(t, "you", p.Values[1])
			assert.NoError(t, p.Err)
		case 1:
			assert.Len(t, p.Values, 1)
			assert.Equal(t, "he", p.Name)
			assert.Equal(t, "too", p.Values[0])
			assert.NoError(t, p.Err)
		}
	}
}

func TestRequestValidateQueryParametersPanic(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 2)

	validator := func(QueryParameter) merry.Error {
		panic(merry.New("i blewed up!"))
	}
	fields := []ParameterSchema{{Name: "zalgo"}, {Name: "waits", Validator: validator}}

	_, _, exception := cut.ValidateQueryParameters(fields)
	assert.Error(t, exception)
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

	validator := func(QueryParameter) merry.Error {
		return merry.New("meh")
	}
	fields := []ParameterSchema{{Name: "zalgo"}, {Name: "waits", Validator: validator}}

	malformed, unknown, exception := cut.ValidateQueryParameters(fields)
	assert.True(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, exception)
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

	validator := func(QueryParameter) merry.Error {
		return merry.New("meh")
	}
	fields := []ParameterSchema{{Name: "zalgo"}, {Name: "waits", Validator: validator}}

	malformed, unknown, exception := cut.ValidateQueryParameters(fields)
	assert.True(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, exception)
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

	fields := []ParameterSchema{{Name: "zalgo"}}

	malformed, unknown, exception := cut.ValidateQueryParameters(fields)
	assert.False(t, malformed)
	assert.True(t, unknown)
	assert.NoError(t, exception)
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

	validator := func(QueryParameter) merry.Error {
		return merry.New("meh")
	}
	fields := []ParameterSchema{{Name: "zalgo", Validator: validator}}

	malformed, unknown, exception := cut.ValidateQueryParameters(fields)
	assert.True(t, malformed)
	assert.True(t, unknown)
	assert.NoError(t, exception)
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

	fields := []ParameterSchema{{Name: "zalgo"}, {Name: "waits", Default: "behind the walls"}}

	malformed, unknown, exception := cut.ValidateQueryParameters(fields)
	assert.False(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, exception)
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

	fields := []ParameterSchema{{Name: "zalgo"}, {Name: "waits", Required: true}}

	malformed, unknown, exception := cut.ValidateQueryParameters(fields)
	assert.True(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, exception)
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

	fields := []ParameterSchema{{Name: "zalgo"}, {Name: "waits"}}

	malformed, unknown, exception := cut.ValidateQueryParameters(fields)
	assert.False(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, exception)
	assert.Len(t, cut.QueryParams, 1)
	assert.NoError(t, cut.QueryParams[0].Err)
}

func TestRequestValidateQueryParameters(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 2)

	fields := []ParameterSchema{{Name: "zalgo"}, {Name: "waits"}}

	malformed, unknown, exception := cut.ValidateQueryParameters(fields)
	assert.False(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, exception)
	assert.NoError(t, cut.QueryParams[0].Err)
	assert.NoError(t, cut.QueryParams[1].Err)
}

func TestRequestValidateQueryParametersNoFields(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 2)

	malformed, unknown, exception := cut.ValidateQueryParameters([]ParameterSchema{})
	assert.False(t, malformed)
	assert.False(t, unknown)
	assert.NoError(t, exception)
}

func TestQueryParamExists(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	assert.True(t, cut.ParseQueryParameters())
	assert.Len(t, cut.QueryParams, 1)

	assert.True(t, cut.QueryParamExists("zalgo"))
	assert.False(t, cut.QueryParamExists("foo"))
}

func TestPathParamExists(t *testing.T) {
	url := "http://example.com/test/thing?zalgo=he:comes"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	cut.PathParams = []PathParameter{{Name: "location", Value: "test"}}

	assert.True(t, cut.PathParamExists("location"))
	assert.False(t, cut.PathParamExists("foo"))
}

func TestRequestID(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}

	id := cut.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, id, cut.ID())
}

func TestRequestClientIP(t *testing.T) {
	url := "http://example.com/test?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, url, nil)
	cut := &Request{Request: request}
	cut.RemoteAddr = "192.168.1.1:65666"

	ip := cut.ClientIP()
	assert.Equal(t, ip, "192.168.1.1")
	assert.Equal(t, ip, cut.ClientIP())
}
