package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func TestCSRFProtector_Service(t *testing.T) {
	defaultSecret := "%wgc83eKEPgdvOBn0NSPG_qsf11VSZLG"
	defaultInvalidSecret := "123483eKEPgdvOBn0NSPG_qsf11VSZLG"
	c := context.New(nil)

	servicetests := []struct {
		headerKeys     []string
		headerVals     []string
		siteurl        string
		token          string
		cookieVal      string
		expectedStatus int
	}{
		// Missing Origin/Referer headers
		{[]string{}, []string{}, "http://example.com", defaultSecret, defaultSecret, http.StatusForbidden},
		// Unparseable SiteUrl
		{[]string{"Origin"}, []string{"http://example.com"}, ":", defaultSecret, defaultSecret, http.StatusInternalServerError},
		// Unparseable Origin
		{[]string{"Origin"}, []string{":"}, "http://example.com", defaultSecret, defaultSecret, http.StatusInternalServerError},
		// Mismatched Origin/SiteUrl
		{[]string{"Origin"}, []string{"http://malicious.com"}, "http://example.com", defaultSecret, defaultSecret, http.StatusForbidden},
		// Unparseable Referer
		{[]string{"Referer"}, []string{":"}, "http://example.com", defaultSecret, defaultSecret, http.StatusInternalServerError},
		// Mismatched Referer/SiteUrl
		{[]string{"Referer"}, []string{"http://malicious.com"}, "http://example.com", defaultSecret, defaultSecret, http.StatusForbidden},
		// Success - Origin header
		{[]string{"Origin"}, []string{"http://example.com"}, "http://example.com", defaultSecret, defaultSecret, 0},
		// Success - Referer header
		{[]string{"Referer"}, []string{"http://example.com"}, "http://example.com", defaultSecret, defaultSecret, 0},
		// No cookie present
		{[]string{"Referer"}, []string{"http://example.com"}, "http://example.com", defaultSecret, "", http.StatusForbidden},
		// Wrong length cookie value
		{[]string{"Referer"}, []string{"http://example.com"}, "http://example.com", defaultSecret, "wronglength", http.StatusForbidden},
		// Error extracting token
		{[]string{"Referer"}, []string{"http://example.com"}, "http://example.com", "", defaultSecret, http.StatusForbidden},
		// Wrong-length token
		{[]string{"Referer"}, []string{"http://example.com"}, "http://example.com", "wronglength", defaultSecret, http.StatusForbidden},
		// Invalid token
		{[]string{"Referer"}, []string{"http://example.com"}, "http://example.com", defaultInvalidSecret, defaultSecret, http.StatusForbidden},
	}

	for _, tt := range servicetests {
		p := CSRFProtector{
			SiteURL:      tt.siteurl,
			ExtractToken: dummyTokenExtractor(tt.token),
		}

		httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
		req := &service.Request{
			Request: httpReq,
		}

		for i, k := range tt.headerKeys {
			req.Header.Add(k, tt.headerVals[i])
		}

		if tt.cookieVal != "" {
			req.AddCookie(&http.Cookie{
				Name:  defaultCookieName,
				Value: tt.cookieVal,
			})
		}

		resp := p.Service(c, req)

		if resp == nil {
			assert.Zerof(t, tt.expectedStatus, "%v response for %v when expected %v", resp, tt, tt.expectedStatus)
		} else {
			assert.Equalf(t, tt.expectedStatus, resp.StatusCode(), "received %v response for %v when expected %v", resp.StatusCode(), tt, tt.expectedStatus)
		}
	}
}

func dummyTokenExtractor(token string) authn.TokenExtractor {
	return func(c context.Context, r *service.Request) (string, merry.Error) {
		if token == "" {
			return token, merry.New("No token")
		}
		return token, nil
	}
}
