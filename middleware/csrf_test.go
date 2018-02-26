package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

const (
	defaultSecret        = "%wgc83eKEPgdvOBn0NSPG_qsf11VSZLG"
	defaultInvalidSecret = "123483eKEPgdvOBn0NSPG_qsf11VSZLG"
	defaultSiteURL       = "http://example.com"
)

type serviceTest struct {
	headerKey      string
	headerVal      string
	siteurl        string
	token          string
	cookieVal      string
	expectedStatus int
}

func checkServiceTest(t *testing.T, c context.Context, st serviceTest) {
	s, err := url.Parse(st.siteurl)
	if err != nil {
		t.Errorf("error parsing url: %v", err)
		return
	}
	p := CSRFProtector{
		SiteURL: *s,
	}

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &httpx.Request{
		Request: httpReq,
	}

	if st.headerKey != "" {
		vals := strings.Split(st.headerVal, ",")
		for _, v := range vals {
			req.Header.Add(st.headerKey, v)
		}
	}
	req.Header.Add("X-CSRF-Token", st.token)

	if st.cookieVal != "" {
		req.AddCookie(&http.Cookie{
			Name:  defaultCookieName,
			Value: st.cookieVal,
		})
	}

	resp := p.Service(c, req)

	if resp == nil {
		assert.Zerof(t, st.expectedStatus, "%v response for %v when expected %v", resp, st, st.expectedStatus)
	} else {
		assert.Equalf(t, st.expectedStatus, resp.StatusCode(), "received %v response for %v when expected %v", resp.StatusCode(), st, st.expectedStatus)
	}
}

func TestCSRFProtector_Service(t *testing.T) {
	c := context.New(nil)

	servicetests := []serviceTest{
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
		checkServiceTest(t, c, tt)
	}

	s, err := url.Parse(defaultSiteURL)
	if err != nil {
		t.Errorf("error parsing url: %v", err)
	}

	// N.B. Check `IsExempt` Hook
	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &httpx.Request{
		Request: httpReq,
	}

	p := CSRFProtector{
		SiteURL: *s,
		IsExempt: func(c context.Context, r *httpx.Request) bool {
			return true
		},
	}
	resp := p.Service(c, req)
	assert.Nil(t, resp, "response should be nil for CSRF exempt request")

	// N.B. Check `TokenExtractor` Hook
	httpReq = httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req = &httpx.Request{
		Request: httpReq,
	}

	req.Header.Add("Origin", defaultSiteURL)
	req.Header.Add("X-CSRF-Token", defaultSecret)
	req.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: defaultSecret,
	})

	p = CSRFProtector{
		SiteURL: *s,
		ExtractToken: func(c context.Context, r *httpx.Request) (string, merry.Error) {
			return "", merry.New("token not found")
		},
	}
	resp = p.Service(c, req)
	if resp == nil {
		assert.NotNil(t, resp, "response status be should 403 for 'Token Not Found', got nil response")
	} else {
		assert.Equal(t, 403, resp.StatusCode(), "response status should be 403 for 'Token Not Found', got %v", resp.StatusCode())
	}

	// N.B. Check `CookieName` Hook
	httpReq = httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req = &httpx.Request{
		Request: httpReq,
	}

	cname := "csrfspecial"
	req.Header.Add("Origin", defaultSiteURL)
	req.Header.Add("X-CSRF-Token", defaultSecret)
	req.AddCookie(&http.Cookie{
		Name:  cname,
		Value: defaultSecret,
	})

	p = CSRFProtector{
		SiteURL:    *s,
		CookieName: cname,
	}
	resp = p.Service(c, req)
	assert.Nil(t, resp, "response should be nil for custom cookie name")

	// N.B. Check `TokenLength` Hook
	httpReq = httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req = &httpx.Request{
		Request: httpReq,
	}

	specialLength := 13
	req.Header.Add("Origin", defaultSiteURL)
	req.Header.Add("X-CSRF-Token", defaultSecret[:specialLength])
	req.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: defaultSecret[:specialLength],
	})

	p = CSRFProtector{
		SiteURL:     *s,
		TokenLength: specialLength,
	}
	resp = p.Service(c, req)
	assert.Nil(t, resp, "response should be nil for custom token length")

	// N.B. Check `CheckOrigin` Hook
	httpReq = httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req = &httpx.Request{
		Request: httpReq,
	}

	req.Header.Add("Origin", defaultSiteURL)
	req.Header.Add("X-CSRF-Token", defaultSecret)
	req.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: defaultSecret,
	})

	p = CSRFProtector{
		SiteURL: *s,
		CheckOrigin: func(expected, actual url.URL) bool {
			return false
		},
	}
	resp = p.Service(c, req)
	if resp == nil {
		assert.NotNil(t, resp, "response status be should 403 for failed origin check, got nil response")
	} else {
		assert.Equal(t, 403, resp.StatusCode(), "response status should be 403 for failed origin check, got %v", resp.StatusCode())
	}

	// N.B. Check `ErrorHandler` Hook
	httpReq = httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req = &httpx.Request{
		Request: httpReq,
	}

	req.Header.Add("Origin", defaultSiteURL)
	req.Header.Add("X-CSRF-Token", defaultSecret)

	p = CSRFProtector{
		SiteURL: *s,
		ErrorHandler: func(c context.Context, r *httpx.Request, err merry.Error) httpx.Response {
			return httpx.NewEmpty(http.StatusTeapot)
		},
	}

	resp = p.Service(c, req)
	if resp == nil {
		assert.NotNil(t, resp, "response status be should 418 for custom error handler, got nil response")
	} else {
		assert.Equal(t, 418, resp.StatusCode(), "response status should be 418 for custom error handler, got %v", resp.StatusCode())
	}
}
