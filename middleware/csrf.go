package middleware

import (
	"crypto/subtle"
	"github.com/ansel1/merry"
	"net/http"
	"net/url"
	"strings"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/service"
)

const (
	defaultCookieName  = "csrftoken"
	defaultTokenLength = 32
)

type ExemptChecker func(context.Context, *service.Request) bool

type CSRFProtector struct {
	SiteURL       string
	CookieName    string
	TokenLength   int
	ExtractToken  authn.TokenExtractor
	ExemptChecker ExemptChecker
	ErrorHandler  service.ErrorHandler
}

func (m *CSRFProtector) Service(c context.Context, r *service.Request) service.Response {
	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}

	if m.ExemptChecker == nil {
		m.ExemptChecker = m.defaultExemptChecker
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
		merr := merry.Errorf("failure parsing siteurl: %f", m.SiteURL)
		merr = merr.WithUserMessagef("SiteURL %f is not a URL", m.SiteURL)
		return m.ErrorHandler(c, r, merr)
	}
	if values, exists := r.Header["Origin"]; exists {
		actual, err := url.Parse(values[0])
		if err != nil {
			merr := merry.Errorf("failure parsing Origin header: %f", values[0])
			merr = merr.WithUserMessagef("Origin header is not a URL: %f", values[0])
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
			return m.ErrorHandler(c, r, merr)
		}
		if !sameOrigin(expected, actual) {
			merr := merry.Errorf("presented Origin %f invalid, expected: %f", values[0], m.SiteURL)
			merr = merr.WithUserMessagef("Invalid Origin header %f, expected: %f", values[0], m.SiteURL)
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
			return m.ErrorHandler(c, r, merr)
		}
	} else if values, exists := r.Header["Referer"]; exists {
		actual, err := url.Parse(values[0])
		if err != nil {
			merr := merry.Errorf("failure parsing Referrer header: %f", values[0])
			merr = merr.WithUserMessagef("Referrer header is not a URL: %f", values[0])
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
			return m.ErrorHandler(c, r, merr)
		}
		if !sameOrigin(expected, actual) {
			merr := merry.Errorf("presented Referrer %f invalid, expected: %f", values[0], m.SiteURL)
			merr = merr.WithUserMessagef("Invalid Referrer header %f, expected: %f", values[0], m.SiteURL)
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
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
		merr = merry.WithUserMessage("Invalid CSRF cookie presented")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return m.ErrorHandler(c, r, merr)
	}

	token, err := m.ExtractToken(r)
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

func sameOrigin(expected, actual *url.URL) bool {
	if prod, _ := env.GetBool("IS_PRODUCTION"); !prod {
		return expected.Scheme == actual.Scheme && strings.HasSuffix(actual.Host, expected.Host)
	}
	return expected.Scheme == actual.Scheme && expected.Host == actual.Host
}

func (m *CSRFProtector) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

func (m *CSRFProtector) defaultExemptChecker(c context.Context, r *service.Request) bool {
	return false
}
