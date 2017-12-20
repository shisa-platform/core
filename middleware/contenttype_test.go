package middleware

import (
	"github.com/percolate/shisa/contenttype"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
	"net/http"
	"testing"
)

func requestWithContentType(method string, c []contenttype.ContentType, t *testing.T) *service.Request {
	httpReq, err := http.NewRequest(method, "http://10.0.0.1/", nil)
	if err != nil {
		t.Fatal("error instantiating request")
	}
	req := &service.Request{
		Request: httpReq,
	}

	for _, cc := range c {
		if method == http.MethodGet && cc.String() != "/" {
			req.Header.Add(acceptHeaderKey, cc.String())
		} else if cc.String() != "/" {
			req.Header.Add(contentTypeHeaderKey, cc.String())
		}
	}
	return req
}

func TestAllowContentTypes_Service(t *testing.T) {
	goodCT := []contenttype.ContentType{*contenttype.DefaultContentType}
	badCT := []contenttype.ContentType{*contenttype.TextPlainContentType}
	wildcardCT := []contenttype.ContentType{contenttype.ContentType{"*", "*"}}
	nilCT := []contenttype.ContentType{contenttype.ContentType{}}
	multivalueCT := []contenttype.ContentType{*contenttype.DefaultContentType, *contenttype.TextPlainContentType}

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
		req := requestWithContentType(tt.method, tt.contentType, t)
		resp := ah.Service(c, req)

		if resp == nil {
			if tt.expectedStatus != 0 {
				t.Errorf("%v response for %v when expected %v", resp, tt, tt.expectedStatus)
			}
		} else {
			if tt.expectedStatus != resp.StatusCode() {
				t.Errorf("received %v response for %v when expected %v", resp.StatusCode(), tt, tt.expectedStatus)
			}
		}
	}
}

func TestRestrictContentTypes_Service(t *testing.T) {
	goodCT := []contenttype.ContentType{*contenttype.DefaultContentType}
	badCT := []contenttype.ContentType{*contenttype.TextPlainContentType}
	wildcardCT := []contenttype.ContentType{contenttype.ContentType{"*", "*"}}
	nilCT := []contenttype.ContentType{contenttype.ContentType{}}
	multivalueCT := []contenttype.ContentType{*contenttype.DefaultContentType, *contenttype.TextPlainContentType}

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
		req := requestWithContentType(tt.method, tt.contentType, t)
		resp := ah.Service(c, req)

		if resp == nil {
			if tt.expectedStatus != 0 {
				t.Errorf("%v response for %v when expected %v", resp, tt, tt.expectedStatus)
			}
		} else {
			if tt.expectedStatus != resp.StatusCode() {
				t.Errorf("received %v response for %v when expected %v", resp.StatusCode(), tt, tt.expectedStatus)
			}
		}
	}
}
