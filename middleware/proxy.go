package middleware

import (
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

var (
	bufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 32*1024)
		},
	}
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
type Route func(context.Context, *service.Request) *service.Request

// Invoke sends the proxied request and returns the response.
type Invoke func(context.Context, *service.Request) (service.Response, merry.Error)

// Respond modifies the response from the proxied server.
type Respond func(context.Context, *service.Request, service.Response) service.Response

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
	ErrorHandler service.ErrorHandler

	tripper http.RoundTripper
}

func (m *ReverseProxy) Service(ctx context.Context, r *service.Request) service.Response {
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

	request := m.Router(ctx, r)
	if request == nil {
		err := merry.New("proxy request is nil")
		err = err.WithUserMessage("middleware.ReverseProxy.Router returned nil request")
		err = err.WithHTTPCode(http.StatusBadGateway)
		return m.ErrorHandler(ctx, r, err)
	}

	if r.ContentLength == 0 {
		request.Body = nil
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

func (m *ReverseProxy) defaultInvoker(ctx context.Context, req *service.Request) (service.Response, merry.Error) {
	if m.tripper == nil {
		m.tripper = http.DefaultTransport
	}

	response, err := m.tripper.RoundTrip(req.Request)
	if err != nil {
		return nil, merry.Wrap(err).WithHTTPCode(http.StatusBadGateway)
	}

	return ProxyResponse{response}, nil
}

func (m *ReverseProxy) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

// ProxyResponse is an adapter for `http.Response` to the
// `service.Response` interface.
type ProxyResponse struct {
	*http.Response
}

func (r ProxyResponse) StatusCode() int {
	return r.Response.StatusCode
}

func (r ProxyResponse) Headers() http.Header {
	return r.Header
}

func (r ProxyResponse) Trailers() http.Header {
	return r.Trailer
}

func (r ProxyResponse) Serialize(w io.Writer) (n int, err error) {
	buf := getBuffer()
	defer putBuffer(buf)

	var nw int64
	nw, err = io.CopyBuffer(w, r.Body, buf)
	n = int(nw)
	r.Body.Close()

	return
}

func getBuffer() []byte {
	buf := bufPool.Get().([]byte)
	buf = buf[:cap(buf)]
	return buf
}

func putBuffer(buf []byte) {
	bufPool.Put(buf)
}
