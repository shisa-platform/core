package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/contenttype"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func requestWithContentType(method string, c []contenttype.ContentType) *service.Request {
	httpReq := httptest.NewRequest(method, "http://10.0.0.1/", nil)
	req := &service.Request{
		Request: httpReq,
	}

	for _, cc := range c {
		if method == http.MethodGet && cc.String() != "/" {
			req.Header.Add(AcceptHeaderKey, cc.String())
		} else if cc.String() != "/" {
			req.Header.Add(contenttype.ContentTypeHeaderKey, cc.String())
		}
	}
	return req
}

func TestAllowContentTypes_Service(t *testing.T) {
	goodCT := []contenttype.ContentType{*contenttype.ApplicationJson}
	badCT := []contenttype.ContentType{*contenttype.TextPlain}
	wildcardCT := []contenttype.ContentType{*contenttype.New("*", "*")}
	nilCT := []contenttype.ContentType{contenttype.ContentType{}}
	multivalueCT := []contenttype.ContentType{*contenttype.ApplicationJson, *contenttype.TextPlain}

	c := context.New(nil)

	ah := AllowContentTypes{
		Permitted: []contenttype.ContentType{
			goodCT[0],
		},
	}

	servicetests := []struct {
		method         string
		contentType    []contenttype.ContentType
		expectedStatus int
	}{
		{http.MethodPost, goodCT, 0},
		{http.MethodPost, badCT, http.StatusUnsupportedMediaType},
		{http.MethodPost, nilCT, http.StatusUnsupportedMediaType},
		{http.MethodPost, multivalueCT, http.StatusUnsupportedMediaType},
		{http.MethodGet, goodCT, 0},
		{http.MethodGet, badCT, http.StatusNotAcceptable},
		{http.MethodGet, nilCT, http.StatusNotAcceptable},
		{http.MethodGet, wildcardCT, 0},
	}

	for _, tt := range servicetests {
		req := requestWithContentType(tt.method, tt.contentType)
		resp := ah.Service(c, req)

		if resp == nil {
			assert.Zerof(t, tt.expectedStatus, "%v response for %v when expected %v", resp, tt, tt.expectedStatus)
		} else {
			assert.Equalf(t, tt.expectedStatus, resp.StatusCode(), "received %v response for %v when expected %v", resp.StatusCode(), tt, tt.expectedStatus)
		}
	}
}

func TestRestrictContentTypes_Service(t *testing.T) {
	goodCT := []contenttype.ContentType{*contenttype.ApplicationJson}
	badCT := []contenttype.ContentType{*contenttype.TextPlain}
	wildcardCT := []contenttype.ContentType{*contenttype.New("*", "*")}
	nilCT := []contenttype.ContentType{contenttype.ContentType{}}
	multivalueCT := []contenttype.ContentType{*contenttype.ApplicationJson, *contenttype.TextPlain}

	c := context.New(nil)

	ah := RestrictContentTypes{
		Forbidden: []contenttype.ContentType{
			badCT[0],
		},
	}

	servicetests := []struct {
		method         string
		contentType    []contenttype.ContentType
		expectedStatus int
	}{
		{http.MethodPost, goodCT, 0},
		{http.MethodPost, badCT, http.StatusUnsupportedMediaType},
		{http.MethodPost, nilCT, http.StatusUnsupportedMediaType},
		{http.MethodPost, multivalueCT, http.StatusUnsupportedMediaType},
		{http.MethodGet, goodCT, 0},
		{http.MethodGet, badCT, http.StatusNotAcceptable},
		{http.MethodGet, nilCT, http.StatusNotAcceptable},
		{http.MethodGet, wildcardCT, 0},
	}

	for _, tt := range servicetests {
		req := requestWithContentType(tt.method, tt.contentType)
		resp := ah.Service(c, req)

		if resp == nil {
			assert.Zerof(t, tt.expectedStatus, "%v response for %v when expected %v", resp, tt, tt.expectedStatus)
		} else {
			assert.Equalf(t, tt.expectedStatus, resp.StatusCode(), "received %v response for %v when expected %v", resp.StatusCode(), tt, tt.expectedStatus)
		}
	}
}
