package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
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

// Router returns the request to contact the proxied server.
type Router func(context.Context, *httpx.Request) (*httpx.Request, merry.Error)

func (r Router) InvokeSafely(ctx context.Context, request *httpx.Request) (out *httpx.Request, err merry.Error, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in router")

	out, err = r(ctx, request)

	return
}

// Invoker sends the proxied request and returns the response.
type Invoker func(context.Context, *httpx.Request) (httpx.Response, merry.Error)

func (i Invoker) InvokeSafely(ctx context.Context, request *httpx.Request) (response httpx.Response, err merry.Error, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in invoker")

	response, err = i(ctx, request)

	return
}

// Responder modifies the response from the proxied server.
type Responder func(context.Context, *httpx.Request, httpx.Response) (httpx.Response, merry.Error)

func (r Responder) InvokeSafely(ctx context.Context, request *httpx.Request, in httpx.Response) (out httpx.Response, err merry.Error, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in responder")

	out, err = r(ctx, request, in)

	return
}

// ReverseProxy is a Handler that takes an incoming request and
// sends it to another server, proxying the response back to the
// user agent.
type ReverseProxy struct {
	// Router must be non-nil or an InternalServiceError
	// status response will be returned.
	Router Router

	// Invoker can be set to optionally customize how the proxied
	// server is contacted.  If this is not set
	// `http.DefaultTransport` will be used.
	Invoker Invoker

	// Responder can be set to optionally customize the response
	// from the proxied server.  If this is not set the response
	// will not be modified.
	Responder Responder

	// ErrorHandler can be set to optionally customize the
	// response for an error. The `err` parameter passed to the
	// handler will have a recommended HTTP status code. The
	// default handler will return the recommended status code
	// and an empty body.
	ErrorHandler httpx.ErrorHandler
}

func (m *ReverseProxy) Service(ctx context.Context, r *httpx.Request) httpx.Response {
	subCtx := ctx
	if ctx.Span() != nil {
		var span opentracing.Span
		span, subCtx = context.StartSpan(ctx, "ReverseHTTPProxy")
		defer span.Finish()
		ext.Component.Set(span, "middleware")
	}

	request := &httpx.Request{Request: r.WithContext(subCtx)}

	request.Header = cloneHeaders(r.Header)
	request.QueryParams = cloneQueryParams(r.QueryParams)
	request.PathParams = clonePathParams(r.PathParams)

	if r.ContentLength == 0 {
		request.Body = nil
	}

	request, response := m.route(subCtx, request)
	if response != nil {
		return response
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

	response = m.invoke(subCtx, request)

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

	return m.respond(subCtx, request, response)
}

func (m *ReverseProxy) route(ctx context.Context, request *httpx.Request) (*httpx.Request, httpx.Response) {
	subCtx := ctx
	if ctx.Span() != nil {
		var span opentracing.Span
		span, subCtx = context.StartSpan(ctx, "ReverseHTTPProxy.Route")
		defer span.Finish()
	}

	if m.Router == nil {
		err := merry.New("proxy middleware: check invariants: router is nil")
		return nil, m.handleError(subCtx, request, err)
	}

	out, err, exception := m.Router.InvokeSafely(subCtx, request)
	if exception != nil {
		exception = exception.Prepend("proxy middleware: run Router")
		return nil, m.handleError(subCtx, request, exception)
	} else if err != nil {
		err = err.Prepend("proxy middleware: run Router")
		err = err.WithHTTPCode(http.StatusBadGateway)
		return nil, m.handleError(subCtx, request, err)
	} else if out == nil {
		err := merry.New("proxy middleware: run Router: result is nil")
		err = err.WithHTTPCode(http.StatusBadGateway)
		return nil, m.handleError(subCtx, request, err)
	}

	return out, nil
}

func (m *ReverseProxy) invoke(ctx context.Context, request *httpx.Request) httpx.Response {
	subCtx := ctx
	span := noopSpan
	if ctx.Span() != nil {
		span, subCtx = context.StartSpan(ctx, "ReverseHTTPProxy.Invoke")
		defer span.Finish()
	}

	carrier := opentracing.HTTPHeadersCarrier(request.Header)
	if err := span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier); err != nil {
		err1 := merry.Prepend(err, "proxy middleware: inject open tracing headers")
		err1 = err1.WithHTTPCode(http.StatusBadGateway)
		return m.handleError(subCtx, request, err1)
	}

	if m.Invoker == nil {
		response, err := http.DefaultTransport.RoundTrip(request.Request)
		if err != nil {
			err1 := merry.Prepend(err, "proxy middleware: run default invoker")
			err1 = err1.WithHTTPCode(http.StatusBadGateway)
			return m.handleError(subCtx, request, err1)
		}

		return httpx.ResponseAdapter{Response: response}
	}

	response, err, exception := m.Invoker.InvokeSafely(subCtx, request)
	if exception != nil {
		exception = exception.Prepend("proxy middleware: run Invoker")
		response = m.handleError(subCtx, request, exception)
	} else if err != nil {
		err = err.Prepend("proxy middleware: run Invoker")
		err = err.WithHTTPCode(http.StatusBadGateway)
		response = m.handleError(subCtx, request, err)
	} else if response == nil {
		err := merry.New("proxy middleware: run Invoker: result is nil")
		err = err.WithHTTPCode(http.StatusBadGateway)
		response = m.handleError(subCtx, request, err)
	}

	return response
}

func (m *ReverseProxy) respond(ctx context.Context, request *httpx.Request, response httpx.Response) httpx.Response {
	subCtx := ctx
	if ctx.Span() != nil {
		var span opentracing.Span
		span, subCtx = context.StartSpan(ctx, "ReverseHTTPProxy.Respond")
		defer span.Finish()
	}

	if m.Responder == nil {
		return response
	}

	out, err, exception := m.Responder.InvokeSafely(subCtx, request, response)
	if exception != nil {
		exception = exception.Prepend("proxy middleware: run Responder")
		out = m.handleError(subCtx, request, exception)
	} else if err != nil {
		err = err.Prepend("proxy middleware: run Responder")
		err = err.WithHTTPCode(http.StatusBadGateway)
		out = m.handleError(subCtx, request, err)
	} else if out == nil {
		err := merry.New("proxy middleware: run Responder: result is nil")
		err = err.WithHTTPCode(http.StatusBadGateway)
		out = m.handleError(subCtx, request, err)
	}

	return out
}

func (m *ReverseProxy) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	span := noopSpan
	if ctxSpan := ctx.Span(); ctxSpan != nil {
		span = ctxSpan
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
	}

	if m.ErrorHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err)
	}

	response, exception := m.ErrorHandler.InvokeSafely(ctx, request, err)
	if exception != nil {
		exception = exception.Prepend("proxy middleware: run ErrorHandler")
		span.LogFields(otlog.String("exception", exception.Error()))
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
	}

	return response
}

func cloneQueryParams(params []*httpx.QueryParameter) []*httpx.QueryParameter {
	p2 := make([]*httpx.QueryParameter, len(params))
	for i, param := range params {
		dup := new(httpx.QueryParameter)
		*dup = *param
		p2[i] = dup
	}

	return p2
}

func clonePathParams(params []httpx.PathParameter) []httpx.PathParameter {
	p2 := make([]httpx.PathParameter, len(params))
	copy(p2, params)

	return p2
}

func cloneHeaders(headers http.Header) http.Header {
	h2 := make(http.Header, len(headers))
	for k, vv := range headers {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}

	return h2
}
