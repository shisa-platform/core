package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

var (
	hopHeaders = []string{
		"Connection",
		"Proxy-Connection", // non-standard, sent by libcurl
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	}
)

// Route returns the request to contact the proxied server.
type Route func(context.Context, *httpx.Request) (*httpx.Request, merry.Error)

// Invoke sends the proxied request and returns the response.
type Invoke func(context.Context, *httpx.Request) (httpx.Response, merry.Error)

// Respond modifies the response from the proxied server.
type Respond func(context.Context, *httpx.Request, httpx.Response) httpx.Response

// ReverseProxy is a Handler that takes an incoming request and
// sends it to another server, proxying the response back to the
// user agent.
type ReverseProxy struct {
	// Router must be non-nil or an InternalServiceError
	// status response will be returned.
	Router Route
	// Invoker can be set to optionally customize how the proxied
	// server is contacted.  If this is not set
	// `http.DefaultTransport` will be used.
	Invoker Invoke
	// Router can be set to optionally customize the response
	// from the proxied server.  If this is not set the response
	// will not be modified.
	Responder Respond
	// ErrorHandler can be set to optionally customize the
	// response for an error. The `err` parameter passed to the
	// handler will have a recommended HTTP status code. The
	// default handler will return the recommended status code
	// and an empty body.
	ErrorHandler httpx.ErrorHandler

	tripper http.RoundTripper
}

func (m *ReverseProxy) Service(ctx context.Context, r *httpx.Request) httpx.Response {
	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}

	if m.Invoker == nil {
		m.Invoker = m.defaultInvoker
	}

	if m.Router == nil {
		err := merry.New("router is nil")
		err = err.WithUserMessage("middleware.ReverseProxy.Router must be non-nil")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return m.ErrorHandler(ctx, r, err)
	}

	request := &httpx.Request{Request: r.WithContext(ctx)}

	request.Header = cloneHeaders(r.Header)
	request.QueryParams = cloneQueryParams(r.QueryParams)
	request.PathParams = clonePathParams(r.PathParams)

	if r.ContentLength == 0 {
		request.Body = nil
	}

	request, err := m.Router(ctx, request)
	if err != nil {
		err = err.WithHTTPCode(http.StatusBadGateway)
		return m.ErrorHandler(ctx, r, err)
	}
	if request == nil {
		err := merry.New("proxy request is nil")
		err = err.WithUserMessage("middleware.ReverseProxy.Router returned nil request")
		err = err.WithHTTPCode(http.StatusBadGateway)
		return m.ErrorHandler(ctx, r, err)
	}

	request.Close = false

	// Remove hop-by-hop headers listed in the "Connection"
	// header of the request.
	// See https://tools.ietf.org/html/rfc2616#section-14.10
	if c := request.Header.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				request.Header.Del(f)
			}
		}
	}

	// Remove hop-by-hop headers in the request.
	// See https://tools.ietf.org/html/rfc2616#section-13.5.1
	for _, h := range hopHeaders {
		delete(request.Header, h)
	}

	if clientIP, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior, ok := request.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		request.Header.Set("X-Forwarded-For", clientIP)
	}

	response, err := m.Invoker(ctx, request)
	if err != nil {
		err = err.WithHTTPCode(http.StatusBadGateway)
		return m.ErrorHandler(ctx, request, err)
	}
	if response == nil {
		err := merry.New("proxy response is nil")
		err = err.WithUserMessage("middleware.ReverseProxy.Invoker returned nil response")
		err = err.WithHTTPCode(http.StatusBadGateway)
		return m.ErrorHandler(ctx, r, err)
	}

	// Remove hop-by-hop headers listed in the "Connection"
	// header of the response.
	// See https://tools.ietf.org/html/rfc2616#section-14.10
	if c := response.Headers().Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				response.Headers().Del(f)
			}
		}
	}

	// Remove hop-by-hop headers in the response.
	// See https://tools.ietf.org/html/rfc2616#section-13.5.1
	for _, h := range hopHeaders {
		delete(response.Headers(), h)
	}

	if m.Responder == nil {
		return response
	}

	response = m.Responder(ctx, request, response)
	if response == nil {
		err := merry.New("proxy response is nil")
		err = err.WithUserMessage("middleware.ReverseProxy.Responder returned nil response")
		err = err.WithHTTPCode(http.StatusBadGateway)
		return m.ErrorHandler(ctx, r, err)
	}
	return response
}

func (m *ReverseProxy) defaultInvoker(ctx context.Context, req *httpx.Request) (httpx.Response, merry.Error) {
	if m.tripper == nil {
		m.tripper = http.DefaultTransport
	}

	response, err := m.tripper.RoundTrip(req.Request)
	if err != nil {
		return nil, merry.Wrap(err).WithHTTPCode(http.StatusBadGateway)
	}

	return httpx.ResponseAdapter{Response: response}, nil
}

func (m *ReverseProxy) defaultErrorHandler(ctx context.Context, r *httpx.Request, err merry.Error) httpx.Response {
	return httpx.NewEmptyError(merry.HTTPCode(err), err)
}

func cloneQueryParams(p []httpx.QueryParameter) []httpx.QueryParameter {
	p2 := make([]httpx.QueryParameter, len(p))
	copy(p2, p)

	return p2
}

func clonePathParams(p []httpx.PathParameter) []httpx.PathParameter {
	p2 := make([]httpx.PathParameter, len(p))
	copy(p2, p)

	return p2
}

func cloneHeaders(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}

	return h2
}
