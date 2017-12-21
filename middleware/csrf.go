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

// ExemptChecker is a function that takes a context and a
// service.Request reference and determines whether the request
// should be exempt from CSRF protextion
type ExemptChecker func(context.Context, *service.Request) bool

// OriginChecker is a function that takes two url.URLs and returns
// a `bool` denoting whether they should be considered the "same"
// for the purposes of CSRF protection
type OriginChecker func(expected, actual *url.URL) bool

// CSRFProtector is middleware used to guard against CSRF attacks.
type CSRFProtector struct {
	// SiteURL is a string denoting the URL to use for CSRF protection
	SiteURL string

	// ExtractToken is an authn.TokenExtractor that returns the CSRF
	ExtractToken authn.TokenExtractor

	// CookieName can be set to optionally customize the cookie name
	CookieName string

	// TokenLength can be set to optionally customize the CSRF token
	TokenLength int

	// ExemptChecker determines whether the request should be exempt
	// from CSRF protection based on context.Context and *service.Request
	ExemptChecker ExemptChecker

	// OriginChecker determines whether the proviced URLs should be
	// considered the "same" for the purposes of CSRF protection
	OriginChecker OriginChecker

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

	if m.ExemptChecker == nil {
		m.ExemptChecker = m.defaultExemptChecker
	}

	if m.OriginChecker == nil {
		m.OriginChecker = m.defaultOriginChecker
	}

	if m.CookieName == "" {
		m.CookieName = defaultCookieName
	}

	if m.TokenLength == 0 {
		m.TokenLength = defaultTokenLength
	}

	if m.ExemptChecker(c, r) {
		return nil
	}

	var err error

	expected, err := url.Parse(m.SiteURL)
	if err != nil {
		merr := merry.Errorf("failure parsing siteurl: %v", m.SiteURL)
		merr = merr.WithUserMessagef("SiteURL %v is not a URL", m.SiteURL)
		return m.ErrorHandler(c, r, merr)
	}
	if values, exists := r.Header["Origin"]; exists {
		actual, err := url.Parse(values[0])
		if err != nil {
			merr := merry.Errorf("failure parsing Origin header: %v", values[0])
			merr = merr.WithUserMessagef("Origin header is not a URL: %v", values[0])
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
			return m.ErrorHandler(c, r, merr)
		}

		if !m.OriginChecker(expected, actual) {
			merr := merry.Errorf("presented Origin %v invalid, expected: %v", values[0], m.SiteURL)
			merr = merr.WithUserMessagef("Invalid Origin header %v, expected: %v", values[0], m.SiteURL)
			merr = merr.WithHTTPCode(http.StatusForbidden)
			return m.ErrorHandler(c, r, merr)
		}
	} else if values, exists := r.Header["Referer"]; exists {
		actual, err := url.Parse(values[0])
		if err != nil {
			merr := merry.Errorf("failure parsing Referrer header: %v", values[0])
			merr = merr.WithUserMessagef("Referrer header is not a URL: %v", values[0])
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
			return m.ErrorHandler(c, r, merr)
		}
		if !m.OriginChecker(expected, actual) {
			merr := merry.Errorf("presented Referrer %v invalid, expected: %v", values[0], m.SiteURL)
			merr = merr.WithUserMessagef("Invalid Referrer header %v, expected: %v", values[0], m.SiteURL)
			merr = merr.WithHTTPCode(http.StatusForbidden)
			return m.ErrorHandler(c, r, merr)
		}
	} else {
		merr := merry.New("no Origin or Referer header presented")
		merr = merr.WithUserMessage("Missing Origin or Referrer header")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}

	cookie, err := r.Cookie(m.CookieName)
	if err != nil {
		merr := merry.Wrap(err)
		merr = merr.WithMessage("no CSRF cookie presented")
		merr = merr.WithUserMessage("Missing Origin or Referrer header")
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
		merr = merr.WithMessage("invalid CSRF token")
		merr = merr.WithUserMessage("Invalid CSRF token")
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

func (m *CSRFProtector) defaultOriginChecker(expected, actual *url.URL) bool {
	return expected.Scheme == actual.Scheme && expected.Host == actual.Host
}

func (m *CSRFProtector) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

func (m *CSRFProtector) defaultExemptChecker(c context.Context, r *service.Request) bool {
	return false
}
