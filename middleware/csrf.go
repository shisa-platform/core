package middleware

import (
	"crypto/subtle"
	"net/http"
	"net/url"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

const (
	defaultCookieName  = "csrftoken"
	defaultTokenLength = 32
)

// IsExempt is a function that takes a context and a
// service.Request reference and determines whether the request
// should be exempt from CSRF protection
type IsExempt func(context.Context, *service.Request) bool

// CheckOrigin is a function that takes two url.URLs and returns
// a `bool` denoting whether they should be considered the "same"
// for the purposes of CSRF protection
type CheckOrigin func(expected, actual url.URL) bool

// CSRFProtector is middleware used to guard against CSRF attacks.
type CSRFProtector struct {
	// SiteURL denotes the URL to use for CSRF protection - must be non-nil
	// and contain non-empty Scheme and Host values
	SiteURL url.URL

	// ExtractToken is an authn.TokenExtractor that returns the CSRF token
	ExtractToken authn.TokenExtractor

	// CookieName can be set to optionally customize the cookie name,
	// defaults to "csrftoken"
	CookieName string

	// TokenLength can be set to optionally customize the expected
	// CSRF token length, defaults to 32
	TokenLength int

	// IsExempt optionally determines whether the request should
	// be exempt, by default returns `false`
	IsExempt IsExempt

	// CheckOrigin determines whether the provided URLs should be
	// considered the "same" for the purposes of CSRF protection.
	// By default ensures that URL Scheme and Host are equal
	CheckOrigin CheckOrigin

	// ErrorHandler can be set to optionally customize the response
	// for an error. The `err` parameter passed to the handler will
	// have a recommended HTTP status code. The default handler will
	// return the recommended status code and an empty body.
	ErrorHandler service.ErrorHandler
}

func (m *CSRFProtector) Service(c context.Context, r *service.Request) service.Response {
	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}

	if m.IsExempt == nil {
		m.IsExempt = m.defaultIsCSRFExempt
	}

	if m.CheckOrigin == nil {
		m.CheckOrigin = m.defaultCheckOrigin
	}

	if m.ExtractToken == nil {
		m.ExtractToken = m.defaultTokenExtractor
	}

	if m.CookieName == "" {
		m.CookieName = defaultCookieName
	}

	if m.TokenLength == 0 {
		m.TokenLength = defaultTokenLength
	}

	if m.SiteURL.String() == "" || m.SiteURL.Host == "" || m.SiteURL.Scheme == "" {
		merr := merry.New("invalid SiteURL")
		merr = merr.WithUserMessage("CSRFProtector.SiteURL must be non-nil and contain Scheme + Host")
		merr = merr.WithHTTPCode(http.StatusInternalServerError)
		return m.ErrorHandler(c, r, merr)
	}

	if m.IsExempt(c, r) {
		return nil
	}

	values, exists := r.Header["Origin"]
	if !exists {
		values, exists = r.Header["Referer"]
	}
	if !exists {
		merr := merry.New("no Origin/Referer header")
		merr = merr.WithUserMessage("Either Origin or Referrer header must be presented")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}
	if len(values) != 1 {
		merr := merry.New("too many Origin/Referer values")
		merr = merr.WithUserMessage("Too many values provided for Origin or Referer header")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}

	actual, err := url.Parse(values[0])
	if err != nil {
		merr := merry.Wrap(err)
		merr = merr.WithUserMessagef("Origin/Referer header is not a URL: %v", values[0])
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}

	if !m.CheckOrigin(m.SiteURL, *actual) {
		merr := merry.Errorf("invalid origin")
		merr = merr.WithUserMessagef("Origin/Referer header invalid: %v, expected: %v", values[0], m.SiteURL)
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}

	cookie, err := r.Cookie(m.CookieName)
	if err != nil {
		merr := merry.Wrap(err)
		merr = merr.WithMessage("no csrf cookie")
		merr = merr.WithUserMessage("CSRF Cookie must be presented")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}
	if len(cookie.Value) != m.TokenLength {
		merr := merry.New("invalid CSRF cookie")
		merr = merr.WithUserMessage("Invalid CSRF cookie presented")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}

	token, err := m.ExtractToken(c, r)
	if err != nil {
		merr := merry.Wrap(err)
		merr = merr.WithUserMessage("Unable to find CSRF token")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}
	if len(token) != m.TokenLength {
		merr := merry.New("invalid CSRF token")
		merr = merr.WithUserMessage("Invalid CSRF token")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}

	if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) != 1 {
		merr := merry.New("invalid CSRF token")
		merr = merr.WithUserMessage("Invalid CSRF token")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}

	return nil
}

func (m *CSRFProtector) defaultCheckOrigin(expected, actual url.URL) bool {
	return expected.Scheme == actual.Scheme && expected.Host == actual.Host
}

func (m *CSRFProtector) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

func (m *CSRFProtector) defaultIsCSRFExempt(c context.Context, r *service.Request) bool {
	return false
}

func (m *CSRFProtector) defaultTokenExtractor(c context.Context, r *service.Request) (token string, err merry.Error) {
	token = r.Header.Get("X-Csrf-Token")
	if token == "" {
		err = merry.New("missing X-Csrf-Token header")
	}
	return
}
