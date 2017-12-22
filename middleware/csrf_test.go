package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
	defaultSiteURL := "http://example.com"

	c := context.New(nil)

	servicetests := []struct {
		headerKey      string
		headerVal      string
		siteurl        string
		token          string
		cookieVal      string
		expectedStatus int
	}{
		// Missing Origin/Referer headers
		{"", "", defaultSiteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		// Nil SiteUrl
		{"Origin", defaultSiteURL, "", defaultSecret, defaultSecret, http.StatusInternalServerError},
		// Unparseable Origin
		{"Origin", ":", defaultSiteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		// Multiple Origin Headers
		{"Origin", "http://example.com,http://malicious.com", defaultSiteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		// Mismatched Origin/SiteUrl
		{"Origin", "http://malicious.com", defaultSiteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		// Unparseable Referer
		{"Referer", ":", defaultSiteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		// Mismatched Referer/SiteUrl
		{"Referer", "http://malicious.com", defaultSiteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		// Success - Origin header
		{"Origin", defaultSiteURL, defaultSiteURL, defaultSecret, defaultSecret, 0},
		// Success - Referer header
		{"Referer", defaultSiteURL, defaultSiteURL, defaultSecret, defaultSecret, 0},
		// No cookie present
		{"Referer", defaultSiteURL, defaultSiteURL, defaultSecret, "", http.StatusForbidden},
		// Wrong length cookie value
		{"Referer", defaultSiteURL, defaultSiteURL, defaultSecret, "wronglength", http.StatusForbidden},
		// Error extracting token
		{"Referer", defaultSiteURL, defaultSiteURL, "", defaultSecret, http.StatusForbidden},
		// Wrong-length token
		{"Referer", defaultSiteURL, defaultSiteURL, "wronglength", defaultSecret, http.StatusForbidden},
		// Invalid token
		{"Referer", defaultSiteURL, defaultSiteURL, defaultInvalidSecret, defaultSecret, http.StatusForbidden},
	}

	for _, tt := range servicetests {
		s, err := url.Parse(tt.siteurl)
		if err != nil {
			t.Errorf("error parsing url: %v", err)
			continue
		}
		p := CSRFProtector{
			SiteURL: *s,
		}

		httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
		req := &service.Request{
			Request: httpReq,
		}

		if tt.headerKey != "" {
			vals := strings.Split(tt.headerVal, ",")
			for _, v := range vals {
				req.Header.Add(tt.headerKey, v)
			}
		}
		req.Header.Add("X-CSRF-Token", tt.token)

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

	epHttpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	epReq := &service.Request{
		Request: epHttpReq,
	}

	s, err := url.Parse(defaultSiteURL)
	if err != nil {
		t.Errorf("error parsing url: %v", err)
	}

	ep := CSRFProtector{
		SiteURL: *s,
		ExemptChecker: func(c context.Context, r *service.Request) bool {
			return true
		},
	}
	resp := ep.Service(c, epReq)
	assert.Nil(t, resp, "response should be nil for CSRF exempt request")
}

func dummyTokenExtractor(token string) authn.TokenExtractor {
	return func(c context.Context, r *service.Request) (string, merry.Error) {
		if token == "" {
			return token, merry.New("No token")
		}
		return token, nil
	}
}
