package middleware

import (
	"net/http"
	"strings"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/shisa-platform/core/contenttype"
	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
)

const (
	AcceptHeaderKey = "Accept"
)

// RestrictContentTypes is middleware to blacklist incoming
// Content-Type and Accept Headers.
type RestrictContentTypes struct {
	// Forbidden is the content types that should be rejected.
	Forbidden []contenttype.ContentType
	// ErrorHandler can be set to optionally customize the
	// response for an error. The `err` parameter passed to the
	// handler will have a recommended HTTP status code. The
	// default handler will return the recommended status code
	// and an empty body.
	ErrorHandler httpx.ErrorHandler
}

func (m *RestrictContentTypes) Service(ctx context.Context, r *httpx.Request) httpx.Response {
	subCtx := ctx
	if ctx.Span() != nil {
		var span opentracing.Span
		span, subCtx = context.StartSpan(ctx, "RestrictContentTypes")
		defer span.Finish()
		ext.Component.Set(span, "middleware")
	}

	var err merry.Error

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		err = m.checkPayload(r)
	case http.MethodGet:
		err = m.checkQuery(r)
	}

	if err != nil {
		return m.handleError(subCtx, r, err)
	}

	return nil
}

func (m *RestrictContentTypes) checkPayload(r *httpx.Request) (err merry.Error) {
	if values, ok := r.Header[contenttype.ContentTypeHeaderKey]; ok {
		if len(values) != 1 {
			err = merry.New("restrict content type middleware: validate content type: too many values")
			err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
			return
		}
		for _, ct := range m.Forbidden {
			if strings.HasPrefix(values[0], ct.String()) {
				err = merry.New("restrict content type middleware: validate content type: forbidden value")
				err = err.WithValue("value", ct.String())
				err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
				return
			}
		}
	} else {
		err = merry.New("restrict content type middleware: find header: missing Content-Type")
		err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
	}

	return
}

func (m *RestrictContentTypes) checkQuery(r *httpx.Request) (err merry.Error) {
	if values, ok := r.Header[AcceptHeaderKey]; ok {
		for _, value := range values {
			for _, mediaRange := range strings.Split(value, ",") {
				mediaRange = strings.TrimSpace(mediaRange)
				if strings.HasPrefix(mediaRange, "*/*") {
					return
				}

				for _, ct := range m.Forbidden {
					if !strings.HasPrefix(mediaRange, ct.String()) {
						return
					}
				}
			}
		}
		err = merry.New("restrict content type middleware: validate content type: forbidden value")
		err = err.WithValue("value", strings.Join(values, ", "))
	} else {
		err = merry.New("restrict content type middleware: find header: missing Accept")
	}

	err = err.WithHTTPCode(http.StatusNotAcceptable)

	return
}

func (m *RestrictContentTypes) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
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
		exception = exception.Prepend("restrict content type middleware: run ErrorHandler")
		span.LogFields(otlog.String("exception", exception.Error()))
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
	}

	return response
}

// AllowContentTypes is middleware to whitelist incoming
// Content-Type and Accept Headers.
type AllowContentTypes struct {
	// Permitted is content types that should be allowed.
	Permitted []contenttype.ContentType
	// ErrorHandler can be set to optionally customize the response
	// for an error. The `err` parameter passed to the handler will
	// have a recommended HTTP status code. The default handler will
	// return the recommended status code and an empty body.
	ErrorHandler httpx.ErrorHandler
}

func (m *AllowContentTypes) Service(ctx context.Context, r *httpx.Request) httpx.Response {
	subCtx := ctx
	if ctx.Span() != nil {
		var span opentracing.Span
		span, subCtx = context.StartSpan(ctx, "AllowContentTypes")
		defer span.Finish()
		ext.Component.Set(span, "middleware")
	}

	var err merry.Error

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		err = m.checkPayload(r)
	case http.MethodGet:
		err = m.checkQuery(r)
	}

	if err != nil {
		return m.handleError(subCtx, r, err)
	}

	return nil
}

func (m *AllowContentTypes) checkPayload(r *httpx.Request) (err merry.Error) {
	if values, ok := r.Header[contenttype.ContentTypeHeaderKey]; ok {
		if len(values) != 1 {
			err = merry.New("allow content type middleware: validate content type: too many values")
			err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
			return
		}
		for _, ct := range m.Permitted {
			if strings.HasPrefix(values[0], ct.String()) {
				return
			}
		}
		err = merry.New("allow content type middleware: validate content type: unsupported value")
		err = err.WithValue("value", values[0])
		err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
		return
	} else {
		err = merry.New("allow content type middleware: find header: missing Content-Type")
		err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
	}

	return
}

func (m *AllowContentTypes) checkQuery(r *httpx.Request) (err merry.Error) {
	if values, ok := r.Header[AcceptHeaderKey]; ok {
		for _, value := range values {
			for _, mediaRange := range strings.Split(value, ",") {
				mediaRange = strings.TrimSpace(mediaRange)
				if strings.HasPrefix(mediaRange, "*/*") {
					return
				}
				for _, ct := range m.Permitted {
					if strings.HasPrefix(mediaRange, ct.String()) {
						return
					}
				}
			}
		}
		err = merry.New("allow content type middleware: validate content type: unsupported value")
		err = err.WithValue("value", strings.Join(values, ", "))
	} else {
		err = merry.New("allow content type middleware: find header: missing Accept")
	}
	err = err.WithHTTPCode(http.StatusNotAcceptable)

	return
}

func (m *AllowContentTypes) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
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
		exception = exception.Prepend("allow content type middleware: run ErrorHandler")
		span.LogFields(otlog.String("exception", exception.Error()))
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
	}

	return response
}
