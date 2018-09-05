package middleware

import (
	stdctx "context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
)

const (
	defaultSecret        = "%wgc83eKEPgdvOBn0NSPG_qsf11VSZLG"
	defaultInvalidSecret = "123483eKEPgdvOBn0NSPG_qsf11VSZLG"
	defaultSiteURL       = "http://example.com"
)

var (
	siteURL = mustParse(defaultSiteURL)
)

type csrfTest struct {
	name           string
	headerKey      string
	headerVal      string
	siteURL        *url.URL
	token          string
	cookieVal      string
	expectedStatus int
}

func (st csrfTest) check(t *testing.T) {
	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}
	ctx := context.New(stdctx.Background())

	if st.headerKey != "" {
		vals := strings.Split(st.headerVal, ",")
		for _, v := range vals {
			request.Header.Add(st.headerKey, v)
		}
	}
	request.Header.Add("X-CSRF-Token", st.token)

	if st.cookieVal != "" {
		request.AddCookie(&http.Cookie{
			Name:  defaultCookieName,
			Value: st.cookieVal,
		})
	}

	cut := CSRFProtector{
		SiteURL: st.siteURL,
	}

	tracer := mocktracer.New()
	span := tracer.StartSpan("test")
	ctx = ctx.WithSpan(span)

	opentracing.SetGlobalTracer(tracer)

	response := cut.Service(ctx, request)

	opentracing.SetGlobalTracer(opentracing.NoopTracer{})

	if response == nil {
		assert.Zero(t, st.expectedStatus, "expected response")
	} else {
		assert.Equal(t, st.expectedStatus, response.StatusCode(), "unexpected status: %v", response)
	}
}

func mustParse(value string) *url.URL {
	parsed, err := url.Parse(value)
	if err != nil {
		panic(err)
	}

	return parsed
}

func TestCSRFProtectorService(t *testing.T) {
	csrfTests := []csrfTest{
		{"missing origin/referer headers", "", "", siteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		{"nil siteurl", "Origin", defaultSiteURL, nil, defaultSecret, defaultSecret, http.StatusInternalServerError},
		{"empty siteurl scheme", "Origin", defaultSiteURL, &url.URL{Host: "example.com", Path: "/"}, defaultSecret, defaultSecret, http.StatusInternalServerError},
		{"empty siteurl host", "Origin", defaultSiteURL, &url.URL{Scheme: "http", Path: "/"}, defaultSecret, defaultSecret, http.StatusInternalServerError},
		{"unparseable origin", "Origin", ":", siteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		{"multiple origin Headers", "Origin", "http://example.com,http://malicious.com", siteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		{"mismatched origin/siteurl", "Origin", "http://malicious.com", siteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		{"unparseable referer", "Referer", ":", siteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		{"mismatched referer/siteurl", "Referer", "http://malicious.com", siteURL, defaultSecret, defaultSecret, http.StatusForbidden},
		{"no cookie present", "Referer", defaultSiteURL, siteURL, defaultSecret, "", http.StatusForbidden},
		{" wrong cookie length", "Referer", defaultSiteURL, siteURL, defaultSecret, "wronglength", http.StatusForbidden},
		{"error extracting token", "Referer", defaultSiteURL, siteURL, "", defaultSecret, http.StatusForbidden},
		{"wrong token length", "Referer", defaultSiteURL, siteURL, "wronglength", defaultSecret, http.StatusForbidden},
		{"invalid token", "Referer", defaultSiteURL, siteURL, defaultInvalidSecret, defaultSecret, http.StatusForbidden},
		{"success - origin header", "Origin", defaultSiteURL, siteURL, defaultSecret, defaultSecret, 0},
		{"success - referer header", "Referer", defaultSiteURL, siteURL, defaultSecret, defaultSecret, 0},
	}

	for _, test := range csrfTests {
		t.Run(test.name, test.check)
	}
}

func TestCSRFProtectorIsExemptHook(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	cut := CSRFProtector{
		SiteURL: siteURL,
		IsExempt: func(c context.Context, r *httpx.Request) bool {
			return true
		},
	}

	response := cut.Service(ctx, request)
	assert.Nil(t, response, "response should be nil for CSRF exempt request")
}

func TestCSRFProtectorIsExemptHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	cut := CSRFProtector{
		SiteURL: siteURL,
		IsExempt: func(c context.Context, r *httpx.Request) bool {
			panic(merry.New("i blewed up!"))
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestCSRFProtectorServiceCheckOriginHook(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	request.Header.Add("Origin", defaultSiteURL)
	request.Header.Add("X-CSRF-Token", defaultSecret)
	request.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: defaultSecret,
	})

	cut := CSRFProtector{
		SiteURL: siteURL,
		CheckOrigin: func(expected, actual *url.URL) bool {
			return false
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusForbidden, response.StatusCode())
}

func TestCSRFProtectorServiceCheckOriginHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	request.Header.Add("Origin", defaultSiteURL)
	request.Header.Add("X-CSRF-Token", defaultSecret)
	request.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: defaultSecret,
	})

	cut := CSRFProtector{
		SiteURL: siteURL,
		CheckOrigin: func(expected, actual *url.URL) bool {
			panic(merry.New("i blewed up!"))
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestCSRFProtectorServiceCheckOriginHookPanicString(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	request.Header.Add("Origin", defaultSiteURL)
	request.Header.Add("X-CSRF-Token", defaultSecret)
	request.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: defaultSecret,
	})

	cut := CSRFProtector{
		SiteURL: siteURL,
		CheckOrigin: func(expected, actual *url.URL) bool {
			panic("i blewed up!")
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestCSRFProtectorTokenExtractorHook(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	request.Header.Add("Origin", defaultSiteURL)
	request.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: "foo",
	})

	cut := CSRFProtector{
		SiteURL: siteURL,
		ExtractToken: func(c context.Context, r *httpx.Request) (string, merry.Error) {
			return "foo", nil
		},
		TokenLength: 3,
	}

	response := cut.Service(ctx, request)
	assert.Nil(t, response)
}

func TestCSRFProtectorTokenExtractorHookError(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	request.Header.Add("Origin", defaultSiteURL)
	request.Header.Add("X-CSRF-Token", defaultSecret)
	request.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: defaultSecret,
	})

	cut := CSRFProtector{
		SiteURL: siteURL,
		ExtractToken: func(c context.Context, r *httpx.Request) (string, merry.Error) {
			return "", merry.New("token not found")
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusForbidden, response.StatusCode())
}

func TestCSRFProtectorTokenExtractorHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	request.Header.Add("Origin", defaultSiteURL)
	request.Header.Add("X-CSRF-Token", defaultSecret)
	request.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: defaultSecret,
	})

	cut := CSRFProtector{
		SiteURL: siteURL,
		ExtractToken: func(c context.Context, r *httpx.Request) (string, merry.Error) {
			panic(merry.New("i blewed up!"))
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestCSRFProtectorServiceCookieNameHook(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	cookieName := "csrfspecial"
	request.Header.Add("Origin", defaultSiteURL)
	request.Header.Add("X-CSRF-Token", defaultSecret)
	request.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: defaultSecret,
	})

	cut := CSRFProtector{
		SiteURL:    siteURL,
		CookieName: cookieName,
	}

	response := cut.Service(ctx, request)
	assert.Nil(t, response)
}

func TestCSRFProtectorServiceTokenLengthHook(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	specialLength := 13
	request.Header.Add("Origin", defaultSiteURL)
	request.Header.Add("X-CSRF-Token", defaultSecret[:specialLength])
	request.AddCookie(&http.Cookie{
		Name:  defaultCookieName,
		Value: defaultSecret[:specialLength],
	})

	cut := CSRFProtector{
		SiteURL:     siteURL,
		TokenLength: specialLength,
	}

	response := cut.Service(ctx, request)
	assert.Nil(t, response)
}

func TestCSRFProtectorServiceErrorHandlerHook(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	request.Header.Add("Origin", defaultSiteURL)
	request.Header.Add("X-CSRF-Token", defaultSecret)

	cut := CSRFProtector{
		SiteURL: siteURL,
		ErrorHandler: func(c context.Context, r *httpx.Request, err merry.Error) httpx.Response {
			return httpx.NewEmpty(http.StatusTeapot)
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusTeapot, response.StatusCode())
}

func TestCSRFProtectorServiceErrorHandlerHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{
		Request: httpReq,
	}

	request.Header.Add("Origin", defaultSiteURL)
	request.Header.Add("X-CSRF-Token", defaultSecret)

	cut := CSRFProtector{
		SiteURL: siteURL,
		ErrorHandler: func(c context.Context, r *httpx.Request, err merry.Error) httpx.Response {
			panic(merry.New("i blewed up!"))
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusForbidden, response.StatusCode())
}
