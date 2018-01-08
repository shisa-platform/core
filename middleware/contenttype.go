package middleware

import (
	"net/http"
	"strings"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/contenttype"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

const (
	ContentTypeHeaderKey = "Content-Type"
	AcceptHeaderKey      = "Accept"
)

// RestrictContentTypes is middleware to blacklist incoming
// Content-Type and Accept Headers.
//
// `Forbidden` is a contenttype.ContentType slice containing content
// types that should be blacklisted.
// `ErrorHandler` can be set to optionally customize the response
// for an error. The `err` parameter passed to the handler will
// have a recommended HTTP status code. The default handler will
// return the recommended status code and an empty body.
type RestrictContentTypes struct {
	Forbidden    []contenttype.ContentType
	ErrorHandler service.ErrorHandler
}

func (m *RestrictContentTypes) Service(c context.Context, r *service.Request) service.Response {
	var err merry.Error

	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		err = m.checkPayload(r)
	case http.MethodGet:
		err = m.checkQuery(r)
	}

	if err != nil {
		return m.ErrorHandler(c, r, err)
	}

	return nil
}

func (m *RestrictContentTypes) checkPayload(r *service.Request) (err merry.Error) {
	if values, ok := r.Header[contentTypeHeaderKey]; ok {
		if len(values) != 1 {
			err = merry.New("too many content types declared")
			err = err.WithUserMessage("Content-Type header must be a single value")
			err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
			return
		}
		for _, ct := range m.Forbidden {
			if strings.HasPrefix(values[0], ct.String()) {
				err = merry.Errorf("unsupported content type: %q", ct)
				err = err.WithUserMessagef("Unsupported Content-Type: %s", ct)
				err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
				return
			}
		}
	} else {
		err = merry.New("content type header not provided")
		err = err.WithUserMessage("Content-Type header must be provided")
		err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
	}
	return
}

func (m *RestrictContentTypes) checkQuery(r *service.Request) (err merry.Error) {
	if values, ok := r.Header[acceptHeaderKey]; ok {
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
		err = err.WithUserMessage("Unsupported Accept header")
	} else {
		err = merry.New("accept header not provided")
		err = err.WithUserMessage("Accept header must be provided")
	}
	err = err.WithHTTPCode(http.StatusNotAcceptable)
	return
}

func (m *RestrictContentTypes) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

// AllowContentTypes is middleware to whitelist incoming
// Content-Type and Accept Headers.
//
// `Permitted` is a contenttype.ContentType slice containing content
// types that should be allowed.
// `ErrorHandler` can be set to optionally customize the response
// for an error. The `err` parameter passed to the handler will
// have a recommended HTTP status code. The default handler will
// return the recommended status code and an empty body.
type AllowContentTypes struct {
	Permitted    []contenttype.ContentType
	ErrorHandler service.ErrorHandler
}

func (m *AllowContentTypes) Service(c context.Context, r *service.Request) service.Response {
	var err merry.Error

	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		err = m.checkPayload(r)
	case http.MethodGet:
		err = m.checkQuery(r)
	}

	if err != nil {
		return m.ErrorHandler(c, r, err)
	}

	return nil
}

func (m *AllowContentTypes) checkPayload(r *service.Request) (err merry.Error) {
	if values, ok := r.Header[contentTypeHeaderKey]; ok {
		if len(values) != 1 {
			err = merry.New("too many content types declared")
			err = err.WithUserMessage("Content-Type header must be a single value")
			err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
			return
		}
		for _, ct := range m.Permitted {
			if strings.HasPrefix(values[0], ct.String()) {
				return
			}
		}
		err = merry.Errorf("unsupported content type: %q", values[0])
		err = err.WithUserMessagef("Unsupported Content-Type: %s", values[0])
		err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
		return
	} else {
		err = merry.New("content-type header not provided")
		err = err.WithUserMessage("Content-Type header must be provided")
		err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
	}
	return
}

func (m *AllowContentTypes) checkQuery(r *service.Request) (err merry.Error) {
	if values, ok := r.Header[acceptHeaderKey]; ok {
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
		err = err.WithUserMessage("Unsupported Accept header")
	} else {
		err = merry.New("accept header not provided")
		err = err.WithUserMessage("Accept header must be provided")
	}
	err = err.WithHTTPCode(http.StatusNotAcceptable)
	return
}

func (m *AllowContentTypes) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}
