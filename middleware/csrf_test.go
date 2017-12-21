package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func TestCSRFProtector_Service(t *testing.T) {
	defaultSecret := "%wgc83eKEPgdvOBn0NSPG_qsf11VSZLG"
	c := context.New(nil)

	servicetests := []struct {
		headerKeys     []string
		headerVals     []string
		siteurl        string
		token          string
		cookieVal      string
		expectedStatus int
	}{
		{[]string{}, []string{}, "invalid", defaultSecret, defaultSecret, http.StatusInternalServerError},
		{[]string{"Origin"}, []string{"invalid"}, "default.com", defaultSecret, defaultSecret, http.StatusInternalServerError},
		{[]string{"Origin"}, []string{"maliciousdefault.com"}, "default.com", defaultSecret, defaultSecret, 0},
		{[]string{"Referer"}, []string{"invalid"}, "default.com", defaultSecret, defaultSecret, http.StatusInternalServerError},
		{[]string{"Referer"}, []string{"maliciousdefault.com"}, "default.com", defaultSecret, defaultSecret, 0},
	}

	for _, tt := range servicetests {
		p := CSRFProtector{
			SiteURL:      tt.siteurl,
			ExtractToken: tokenExtractorFactory(tt.token),
		}

		httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
		req := &service.Request{
			Request: httpReq,
		}

		for i, k := range tt.headerKeys {
			req.Header.Add(k, tt.headerVals[i])
		}

		req.AddCookie(&http.Cookie{
			Name:  p.CookieName,
			Value: tt.cookieVal,
		})

		resp := p.Service(c, req)

		if resp == nil {
			assert.Zerof(t, tt.expectedStatus, "%v response for %v when expected %v", resp, tt, tt.expectedStatus)
		} else {
			assert.Equalf(t, tt.expectedStatus, resp.StatusCode(), "received %v response for %v when expected %v", resp.StatusCode(), tt, tt.expectedStatus)
		}
	}
}

func tokenExtractorFactory(token string) authn.TokenExtractor {
	return func(c context.Context, r *service.Request) (string, merry.Error) {
		if token == "" {
			return token, merry.New("No token")
		}
		return token
	}
}
