package middleware

import (
	stdctx "context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/contenttype"
	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
)

type contentTypeTest struct {
	name           string
	handler        httpx.Handler
	method         string
	contentType    []contenttype.ContentType
	expectedStatus int
}

var (
	good       = []contenttype.ContentType{*contenttype.ApplicationJson}
	bad        = []contenttype.ContentType{*contenttype.TextPlain}
	wildcard   = []contenttype.ContentType{*contenttype.New("*", "*")}
	empty      = []contenttype.ContentType{contenttype.ContentType{}}
	multivalue = []contenttype.ContentType{*contenttype.ApplicationJson, *contenttype.TextPlain}
)

func (st contentTypeTest) check(t *testing.T) {
	request := requestWithContentType(st.method, st.contentType)
	ctx := context.New(stdctx.Background())

	tracer := mocktracer.New()
	span := tracer.StartSpan("test")
	ctx = ctx.WithSpan(span)

	opentracing.SetGlobalTracer(tracer)

	response := st.handler(ctx, request)

	opentracing.SetGlobalTracer(opentracing.NoopTracer{})

	if response == nil {
		assert.Zero(t, st.expectedStatus)
	} else {
		assert.Equal(t, st.expectedStatus, response.StatusCode(), "unexpected status: %v", response)
	}
}

func requestWithContentType(method string, types []contenttype.ContentType) *httpx.Request {
	httpReq := httptest.NewRequest(method, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	for _, ct := range types {
		value := ct.String()
		if method == http.MethodGet && value != "/" {
			request.Header.Add(AcceptHeaderKey, value)
		} else if value != "/" {
			request.Header.Add(contenttype.ContentTypeHeaderKey, value)
		}
	}

	return request
}

func TestAllowContentTypesService(t *testing.T) {
	allow := AllowContentTypes{
		Permitted: []contenttype.ContentType{
			good[0],
		},
	}

	serviceTests := []contentTypeTest{
		{"post/good", allow.Service, http.MethodPost, good, 0},
		{"post/bad", allow.Service, http.MethodPost, bad, http.StatusUnsupportedMediaType},
		{"post/empty", allow.Service, http.MethodPost, empty, http.StatusUnsupportedMediaType},
		{"post/multi", allow.Service, http.MethodPost, multivalue, http.StatusUnsupportedMediaType},
		{"get/good", allow.Service, http.MethodGet, good, 0},
		{"get/bad", allow.Service, http.MethodGet, bad, http.StatusNotAcceptable},
		{"get/empty", allow.Service, http.MethodGet, empty, http.StatusNotAcceptable},
		{"get/wildcard", allow.Service, http.MethodGet, wildcard, 0},
	}

	for _, test := range serviceTests {
		t.Run(test.name, test.check)
	}
}

func TestAllowContentTypesServiceServiceErrorHandlerHook(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := requestWithContentType(http.MethodGet, bad)

	cut := AllowContentTypes{
		Permitted: []contenttype.ContentType{
			good[0],
		},
		ErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			return httpx.NewEmpty(http.StatusTeapot)
		},
	}

	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusTeapot, response.StatusCode())
}

func TestAllowContentTypesServiceErrorHandlerHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := requestWithContentType(http.MethodGet, bad)

	cut := AllowContentTypes{
		Permitted: []contenttype.ContentType{
			good[0],
		},
		ErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			panic(merry.New("i blewed up!"))
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotAcceptable, response.StatusCode())
}

func TestRestrictContentTypesService(t *testing.T) {
	restrict := RestrictContentTypes{
		Forbidden: []contenttype.ContentType{
			bad[0],
		},
	}

	serviceTests := []contentTypeTest{
		{"post/good", restrict.Service, http.MethodPost, good, 0},
		{"post/bad", restrict.Service, http.MethodPost, bad, http.StatusUnsupportedMediaType},
		{"post/empty", restrict.Service, http.MethodPost, empty, http.StatusUnsupportedMediaType},
		{"post/multi", restrict.Service, http.MethodPost, multivalue, http.StatusUnsupportedMediaType},
		{"get/good", restrict.Service, http.MethodGet, good, 0},
		{"get/bad", restrict.Service, http.MethodGet, bad, http.StatusNotAcceptable},
		{"get/empty", restrict.Service, http.MethodGet, empty, http.StatusNotAcceptable},
		{"get/wildcard", restrict.Service, http.MethodGet, wildcard, 0},
	}

	for _, test := range serviceTests {
		t.Run(test.name, test.check)
	}
}

func TestRestrictContentTypesServiceServiceErrorHandlerHook(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := requestWithContentType(http.MethodGet, bad)

	cut := RestrictContentTypes{
		Forbidden: []contenttype.ContentType{
			bad[0],
		},
		ErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			return httpx.NewEmpty(http.StatusTeapot)
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusTeapot, response.StatusCode())
}

func TestRestrictContentTypesServiceErrorHandlerHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := requestWithContentType(http.MethodGet, bad)

	cut := RestrictContentTypes{
		Forbidden: []contenttype.ContentType{
			bad[0],
		},
		ErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			panic(merry.New("i blewed up!"))
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotAcceptable, response.StatusCode())
}
