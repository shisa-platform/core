package middleware

import (
	"crypto/subtle"
	"net/http"
	"net/url"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

const (
	defaultCookieName  = "csrftoken"
	defaultTokenLength = 32
)

// RequestPredicate examines the given context and request and
// returns a determination based on that analysis.
type RequestPredicate func(context.Context, *httpx.Request) bool

// CheckOrigin compares two URLs and determines if they should be
// considered the "same" for the purposes of CSRF protection.
type CheckOrigin func(expected, actual url.URL) bool

// CSRFProtector is middleware used to guard against CSRF attacks.
type CSRFProtector struct {
	// SiteURL is the URL to use for CSRF protection. This must
	// be non-nil and contain non-empty Scheme and Host values
	// or a internal server error will be returned.
	SiteURL url.URL

	// ExtractToken optionally customizes how the CSRF token is
	// extracted from the request.
	// The default extractor uses the header "X-Csrf-Token".
	ExtractToken httpx.StringExtractor

	// CookieName optionally customizes the name of the CSRF
	// cookie sent by the user agent.
	// The default cookie name is "csrftoken".
	CookieName string

	// TokenLength optionally customizes the expected CSRF token
	// length.
	// The default length is 32.
	TokenLength int

	// IsExempt optionally customizes checking request exemption
	// from CSRF protection.
	// The default checker always returns `false`.
	IsExempt RequestPredicate

	// CheckOrigin optionally customizes how URLs should be
	// compared for the purposes of CSRF protection.
	// The default comparisons ensures that URL Schemes and Hosts
	// are equal.
	CheckOrigin CheckOrigin

	// ErrorHandler optionally customizes the response for an
	// error. The `err` parameter passed to the handler will
	// have a recommended HTTP status code.
	// The default handler will return the recommended status
	// code and an empty body.
	ErrorHandler httpx.ErrorHandler
}

func (m *CSRFProtector) Service(c context.Context, r *httpx.Request) httpx.Response {
	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}

	if m.IsExempt == nil {
		m.IsExempt = m.defaultIsExempt
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

func (m *CSRFProtector) defaultErrorHandler(ctx context.Context, r *httpx.Request, err merry.Error) httpx.Response {
	return httpx.NewEmptyError(merry.HTTPCode(err), err)
}

func (m *CSRFProtector) defaultIsExempt(c context.Context, r *httpx.Request) bool {
	return false
}

func (m *CSRFProtector) defaultTokenExtractor(c context.Context, r *httpx.Request) (token string, err merry.Error) {
	token = r.Header.Get("X-Csrf-Token")
	if token == "" {
		err = merry.New("missing X-Csrf-Token header")
	}
	return
}
