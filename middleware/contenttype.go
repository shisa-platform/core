package middleware

import (
	"net/http"
	"strings"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/contenttype"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
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

func (m *RestrictContentTypes) Service(c context.Context, r *httpx.Request) httpx.Response {
	var err merry.Error

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		err = m.checkPayload(r)
	case http.MethodGet:
		err = m.checkQuery(r)
	}

	if err != nil {
		return m.handleError(c, r, err)
	}

	return nil
}

func (m *RestrictContentTypes) checkPayload(r *httpx.Request) (err merry.Error) {
	if values, ok := r.Header[contenttype.ContentTypeHeaderKey]; ok {
		if len(values) != 1 {
			err = merry.New("too many content types declared")
			err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
			return
		}
		for _, ct := range m.Forbidden {
			if strings.HasPrefix(values[0], ct.String()) {
				err = merry.Errorf("unsupported content type: %q", ct)
				err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
				return
			}
		}
	} else {
		err = merry.New("content type header not provided")
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
		err = merry.Errorf("unsupported accept header: %q", values)
	} else {
		err = merry.New("accept header not provided")
	}
	err = err.WithHTTPCode(http.StatusNotAcceptable)

	return
}

func (m *RestrictContentTypes) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	if m.ErrorHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err)
	}

	var exception merry.Error
	response := m.ErrorHandler.InvokeSafely(ctx, request, err, &exception)
	if exception != nil {
		exception = exception.Prepend("running restrict content type middleware ErrorHandler")
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

func (m *AllowContentTypes) Service(c context.Context, r *httpx.Request) httpx.Response {
	var err merry.Error

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		err = m.checkPayload(r)
	case http.MethodGet:
		err = m.checkQuery(r)
	}

	if err != nil {
		return m.handleError(c, r, err)
	}

	return nil
}

func (m *AllowContentTypes) checkPayload(r *httpx.Request) (err merry.Error) {
	if values, ok := r.Header[contenttype.ContentTypeHeaderKey]; ok {
		if len(values) != 1 {
			err = merry.New("too many content types declared")
			err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
			return
		}
		for _, ct := range m.Permitted {
			if strings.HasPrefix(values[0], ct.String()) {
				return
			}
		}
		err = merry.Errorf("unsupported content type: %q", values[0])
		err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
		return
	} else {
		err = merry.New("content-type header not provided")
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
		err = merry.Errorf("unsupported accept header: %q", values)
	} else {
		err = merry.New("accept header not provided")
	}
	err = err.WithHTTPCode(http.StatusNotAcceptable)

	return
}

func (m *AllowContentTypes) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	if m.ErrorHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err)
	}

	var exception merry.Error
	response := m.ErrorHandler.InvokeSafely(ctx, request, err, &exception)
	if exception != nil {
		exception = exception.Prepend("running allow content type middleware ErrorHandler")
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
	}

	return response
}
