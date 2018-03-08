package middleware

import (
	"fmt"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

func TestReverseProxyMissingRouter(t *testing.T) {
	cut := ReverseProxy{}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.NotNil(t, cut.ErrorHandler)
	assert.NotNil(t, cut.Invoker)
	assert.Nil(t, cut.Responder)
	assert.NotNil(t, cut.ErrorHandler)
}

func TestReverseProxyMissingRouterCustomErrorHandler(t *testing.T) {
	var errorHandlerInvoked bool
	cut := ReverseProxy{
		ErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			errorHandlerInvoked = true
			return httpx.NewEmpty(http.StatusPaymentRequired)
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.True(t, errorHandlerInvoked)
	assert.Equal(t, http.StatusPaymentRequired, response.StatusCode())
}

func TestReverseProxyNilRouterResponse(t *testing.T) {
	var routerInvoked bool
	cut := ReverseProxy{
		Router: func(context.Context, *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			return nil, nil
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.True(t, routerInvoked)
	assert.Equal(t, http.StatusBadGateway, response.StatusCode())
}

func TestReverseProxyNilRouterResponseCustomErrorHandler(t *testing.T) {
	var errorHandlerInvoked bool
	var routerInvoked bool
	cut := ReverseProxy{
		Router: func(context.Context, *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			return nil, nil
		},
		ErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			errorHandlerInvoked = true
			return httpx.NewEmpty(http.StatusPaymentRequired)
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.True(t, routerInvoked)
	assert.True(t, errorHandlerInvoked)
	assert.Equal(t, http.StatusPaymentRequired, response.StatusCode())
}

func TestReverseProxyErrorRouterResponse(t *testing.T) {
	var routerInvoked bool
	cut := ReverseProxy{
		Router: func(context.Context, *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			return nil, merry.New("i blewed up!")
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.True(t, routerInvoked)
	assert.Equal(t, http.StatusBadGateway, response.StatusCode())
}

func TestReverseProxyErrorRouterResponseCustomErrorHandler(t *testing.T) {
	var errorHandlerInvoked bool
	var routerInvoked bool
	cut := ReverseProxy{
		Router: func(context.Context, *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			return nil, merry.New("i blewed up!")
		},
		ErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			errorHandlerInvoked = true
			return httpx.NewEmpty(http.StatusPaymentRequired)
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.True(t, routerInvoked)
	assert.True(t, errorHandlerInvoked)
	assert.Equal(t, http.StatusPaymentRequired, response.StatusCode())
}

func TestReverseProxyInvokerError(t *testing.T) {
	var routerInvoked bool
	var invokerInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			return r, nil
		},
		Invoker: func(c context.Context, r *httpx.Request) (httpx.Response, merry.Error) {
			invokerInvoked = true
			assert.Nil(t, r.Body)
			assert.False(t, r.Close)
			return nil, merry.New("i blewed up!")
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadGateway, response.StatusCode())
	assert.True(t, routerInvoked)
	assert.True(t, invokerInvoked)
}

func TestReverseProxyInvokerErrorCustomErrorHandler(t *testing.T) {
	var errorHandlerInvoked bool
	var routerInvoked bool
	var invokerInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			return r, nil
		},
		Invoker: func(c context.Context, r *httpx.Request) (httpx.Response, merry.Error) {
			invokerInvoked = true
			assert.Nil(t, r.Body)
			assert.False(t, r.Close)
			return nil, merry.New("i blewed up!")
		},
		ErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			errorHandlerInvoked = true
			return httpx.NewEmpty(http.StatusPaymentRequired)
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusPaymentRequired, response.StatusCode())
	assert.True(t, routerInvoked)
	assert.True(t, invokerInvoked)
	assert.True(t, errorHandlerInvoked)
}

func TestReverseProxyNilInvokerResponse(t *testing.T) {
	var routerInvoked bool
	var invokerInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			return r, nil
		},
		Invoker: func(c context.Context, r *httpx.Request) (httpx.Response, merry.Error) {
			invokerInvoked = true
			assert.Nil(t, r.Body)
			assert.False(t, r.Close)
			return nil, nil
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadGateway, response.StatusCode())
	assert.True(t, routerInvoked)
	assert.True(t, invokerInvoked)
}

func TestReverseProxyNilInvokerResponseCustomErrorHandler(t *testing.T) {
	var errorHandlerInvoked bool
	var routerInvoked bool
	var invokerInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			return r, nil
		},
		Invoker: func(c context.Context, r *httpx.Request) (httpx.Response, merry.Error) {
			invokerInvoked = true
			assert.Nil(t, r.Body)
			assert.False(t, r.Close)
			return nil, nil
		},
		ErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			errorHandlerInvoked = true
			return httpx.NewEmpty(http.StatusPaymentRequired)
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusPaymentRequired, response.StatusCode())
	assert.True(t, routerInvoked)
	assert.True(t, invokerInvoked)
	assert.True(t, errorHandlerInvoked)
}

func TestReverseProxySanitizeRequestHeaders(t *testing.T) {
	var routerInvoked bool
	var invokerInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			assert.NotEmpty(t, r.Header.Get("Keep-Alive"))
			assert.NotEmpty(t, r.Header.Get("Upgrade"))
			assert.NotEmpty(t, r.Header.Get("Transfer-Encoding"))
			return r, nil
		},
		Invoker: func(c context.Context, r *httpx.Request) (httpx.Response, merry.Error) {
			invokerInvoked = true
			assert.Nil(t, r.Body)
			assert.False(t, r.Close)
			assert.Empty(t, r.Header.Get("Keep-Alive"))
			assert.Empty(t, r.Header.Get("Upgrade"))
			assert.Empty(t, r.Header.Get("Transfer-Encoding"))
			assert.Empty(t, r.Header.Get("Connection"))
			assert.Empty(t, r.Header.Get("Zalgo"))
			assert.True(t, strings.HasPrefix(r.Header.Get("X-Forwarded-For"), "10.10.10.10, "))
			return httpx.NewEmpty(http.StatusOK), nil
		},
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	request := &httpx.Request{Request: r}
	request.Header.Set("Connection", "keep-alive,zalgo")
	request.Header.Set("Keep-Alive", "true")
	request.Header.Set("Zalgo", "he comes")
	request.Header.Set("Upgrade", "IRC/6.9, zalgo/666")
	request.Header.Set("Transfer-Encoding", "bitrot")
	request.Header.Set("X-Forwarded-For", "10.10.10.10")
	ctx := context.NewFakeContextDefaultFatal(t)

	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode())
	assert.True(t, routerInvoked)
	assert.True(t, invokerInvoked)
}

func TestReverseProxySanitizeResponseHeaders(t *testing.T) {
	var routerInvoked bool
	var invokerInvoked bool
	var responderInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			return r, nil
		},
		Invoker: func(c context.Context, r *httpx.Request) (httpx.Response, merry.Error) {
			invokerInvoked = true
			assert.Nil(t, r.Body)
			assert.False(t, r.Close)
			response := httpx.NewEmpty(http.StatusOK)
			response.Headers().Set("Connection", "keep-alive,zalgo")
			response.Headers().Set("Keep-Alive", "true")
			response.Headers().Set("Zalgo", "he comes")
			response.Headers().Set("Upgrade", "IRC/6.9, zalgo/666")
			response.Headers().Set("Transfer-Encoding", "bitrot")
			return response, nil
		},
		Responder: func(c context.Context, req *httpx.Request, res httpx.Response) httpx.Response {
			responderInvoked = true
			assert.Empty(t, res.Headers().Get("Keep-Alive"))
			assert.Empty(t, res.Headers().Get("Upgrade"))
			assert.Empty(t, res.Headers().Get("Transfer-Encoding"))
			assert.Empty(t, res.Headers().Get("Connection"))
			assert.Empty(t, res.Headers().Get("Zalgo"))
			res.Headers().Set("X-Zalgo", "he comes")
			return res
		},
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	request := &httpx.Request{Request: r}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode())
	assert.True(t, routerInvoked)
	assert.True(t, invokerInvoked)
	assert.True(t, responderInvoked)
	assert.Equal(t, "he comes", response.Headers().Get("X-Zalgo"))
}

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\": \"hello\"}"))

	if f, impl := w.(http.Flusher); impl {
		f.Flush()
	}
}

func TestReverseProxyUncontactableProxy(t *testing.T) {
	var routerInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			r.URL.Scheme = "http"
			r.URL.Host = "zalgo:666"
			r.URL.Path = "/"
			return r, nil
		},
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	request := &httpx.Request{Request: r}
	ctx := context.New(r.Context())

	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadGateway, response.StatusCode())
	assert.True(t, routerInvoked)

	var buf bytes.Buffer
	err := response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, buf.Len())
}

func TestReverseProxyNilResponderResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(helloWorldHandler))
	defer server.Close()

	url, err := url.ParseRequestURI(server.URL)
	if err != nil {
		panic(err)
	}

	var routerInvoked bool
	var responderInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			r.URL.Scheme = url.Scheme
			r.URL.Host = url.Host
			r.URL.Path = "/"
			return r, nil
		},
		Responder: func(c context.Context, req *httpx.Request, res httpx.Response) httpx.Response {
			responderInvoked = true
			return nil
		},
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	request := &httpx.Request{Request: r}
	ctx := context.New(r.Context())

	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadGateway, response.StatusCode())
	assert.True(t, routerInvoked)
	assert.True(t, responderInvoked)

	var buf bytes.Buffer
	err = response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, buf.Len())
}

func TestReverseProxyWithQueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(helloWorldHandler))
	defer server.Close()

	url, err := url.ParseRequestURI(server.URL)
	if err != nil {
		panic(err)
	}

	var routerInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			assert.Len(t, r.QueryParams, 2)
			assert.Equal(t, "zalgo", r.QueryParams[0].Name)
			assert.Len(t, r.QueryParams[0].Values, 1)
			assert.Equal(t, "answer", r.QueryParams[1].Name)
			assert.Len(t, r.QueryParams[1].Values, 1)

			param := r.QueryParams[0]
			r.URL.Scheme = url.Scheme
			r.URL.Host = url.Host
			r.URL.Path = "/"
			r.URL.RawQuery = fmt.Sprintf("%s=%s", param.Name, param.Values[0])
			return r, nil
		},
	}

	r := httptest.NewRequest(http.MethodGet, "/?zalgo=he:comes&answer=42", nil)
	request := &httpx.Request{Request: r}
	assert.True(t, request.ParseQueryParameters())
	ctx := context.New(r.Context())

	response := cut.Service(ctx, request)

	assert.Len(t, request.QueryParams, 2)
	assert.Equal(t, "zalgo", request.QueryParams[0].Name)
	assert.Equal(t, "answer", request.QueryParams[1].Name)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode())
	assert.True(t, routerInvoked)
	assert.Empty(t, response.Trailers())
	assert.NoError(t, response.Err())

	var buf bytes.Buffer
	err = response.Serialize(&buf)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, buf.Len())
	expectedJson := `{
  "message": "hello"
}`
	assert.JSONEq(t, expectedJson, buf.String())
}

func TestReverseProxy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(helloWorldHandler))
	defer server.Close()

	url, err := url.ParseRequestURI(server.URL)
	if err != nil {
		panic(err)
	}

	var routerInvoked bool
	cut := ReverseProxy{
		Router: func(c context.Context, r *httpx.Request) (*httpx.Request, merry.Error) {
			routerInvoked = true
			r.URL.Scheme = url.Scheme
			r.URL.Host = url.Host
			r.URL.Path = "/"
			return r, nil
		},
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	request := &httpx.Request{Request: r}
	ctx := context.New(r.Context())

	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode())
	assert.True(t, routerInvoked)
	assert.Empty(t, response.Trailers())
	assert.NoError(t, response.Err())

	var buf bytes.Buffer
	err = response.Serialize(&buf)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, buf.Len())
	expectedJson := `{
  "message": "hello"
}`
	assert.JSONEq(t, expectedJson, buf.String())
}
