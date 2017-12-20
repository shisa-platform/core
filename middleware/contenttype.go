package middleware

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/contenttype"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
	"strings"
)

var (
	contentTypeHeaderKey = http.CanonicalHeaderKey("Content-Type")
	acceptHeaderKey      = http.CanonicalHeaderKey("Accept")
)

type RestrictContentTypes struct {
	Forbidden    []contenttype.ContentType
	ErrorHandler service.ErrorHandler
}

func (m *RestrictContentTypes) Service(c context.Context, r *service.Request) service.Response {
	var err merry.Error

	if m.ErrorHandler == nil {
		m.ErrorHandler = defaultErrorHandler
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		err = m.checkPayloadBlacklist(r)
	case http.MethodGet:
		err = m.checkQueryBlacklist(r)
	}

	if err != nil {
		return m.ErrorHandler(c, r, err)
	}

	return nil
}

func (m *RestrictContentTypes) checkPayloadBlacklist(r *service.Request) (err merry.Error) {
	if values, ok := r.Header[contentTypeHeaderKey]; ok {
		if len(values) != 1 {
			err = merry.New("Request body Content-Type header must be a single value")
			err.WithHTTPCode(http.StatusUnsupportedMediaType)
			return
		}
		for _, ct := range m.Forbidden {
			if strings.HasPrefix(values[0], ct.String()) {
				err = merry.Errorf("Invalid Content-Type header: %q", ct)
				err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
				return
			}
		}
	} else {
		err = merry.New("Content-Type header must be provided.")
		err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
	}
	return
}

func (m *RestrictContentTypes) checkQueryBlacklist(r *service.Request) (err merry.Error) {
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
		err = merry.New("No valid Accept Content-Type found")
		err = err.WithHTTPCode(http.StatusNotAcceptable)
	} else {
		err = merry.New("No Accept Content-Type provided")
		err = err.WithHTTPCode(http.StatusNotAcceptable)
	}
	return
}

type AllowContentTypes struct {
	Permitted    []contenttype.ContentType
	ErrorHandler service.ErrorHandler
}

func (m *AllowContentTypes) Service(c context.Context, r *service.Request) service.Response {
	var err merry.Error

	if m.ErrorHandler == nil {
		m.ErrorHandler = defaultErrorHandler
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		err = m.checkPayloadWhitelist(r)
	case http.MethodGet:
		err = m.checkQueryWhitelist(r)
	}

	if err != nil {
		return m.ErrorHandler(c, r, err)
	}

	return nil
}

func (m *AllowContentTypes) checkPayloadWhitelist(r *service.Request) (err merry.Error) {
	if values, ok := r.Header[contentTypeHeaderKey]; ok {
		if len(values) != 1 {
			err = merry.New("Request body Content-Type header must be a single value")
			err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
			return
		}
		for _, ct := range m.Permitted {
			if strings.HasPrefix(values[0], ct.String()) {
				return
			}
		}
	} else {
		err = merry.New("Content-Type header must be provided.")
		err = err.WithHTTPCode(http.StatusUnsupportedMediaType)
	}
	return
}
func (m *AllowContentTypes) checkQueryWhitelist(r *service.Request) (err merry.Error) {
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
		err = merry.New("No valid Accept Content-Type header found")
	} else {
		err = merry.New("No Accept Content-Type header provided")
	}
	err = err.WithHTTPCode(http.StatusNotAcceptable)
	return
}

// TODO: move to base middleware file
func defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}
