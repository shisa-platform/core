package middleware

import (
	"crypto/subtle"
	"github.com/ansel1/merry"
	"net/http"
	"net/url"
	"strings"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/service"
)

const (
	defaultCookieName  = "csrftoken"
	defaultTokenLength = 32
)

type CSRFProtector struct {
	CookieName   string
	TokenLength  int
	ExtractToken func(*service.Request) ([]byte, error)
	Exempt       func(context.Context, *service.Request) bool
	ErrorHandler service.ErrorHandler
}

func (p *CSRFProtector) Service(c context.Context, r *service.Request) service.Response {
	if p.ErrorHandler == nil {
		p.ErrorHandler = p.defaultErrorHandler
	}

	if p.CookieName == "" {
		p.CookieName = defaultCookieName
	}

	if p.TokenLength == 0 {
		p.TokenLength = defaultTokenLength
	}

	if p.Exempt(c, r) {
		return nil
	}

	var err error

	siteUrl, merr := env.Get("SITE_URL")
	if merr != nil {
		merr := merry.Wrap(merr)
		return p.ErrorHandler(c, r, merr)
	}
	expected, err := url.Parse(siteUrl)
	if err != nil {
		merr := merry.Errorf("SiteURL %q is not a URL", siteUrl)
		return p.ErrorHandler(c, r, merr)
	}
	if values, exists := r.Header["Origin"]; exists {
		actual, err := url.Parse(values[0])
		if err != nil {
			merr := merry.Errorf("presented Origin %q is not a URL", values[0])
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
			return p.ErrorHandler(c, r, merr)
		}
		if !sameOrigin(expected, actual) {
			merr := merry.Errorf("presented Origin %q invalid, expected: %q", values[0], siteUrl)
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
			return p.ErrorHandler(c, r, merr)
		}
	} else if values, exists := r.Header["Referer"]; exists {
		actual, err := url.Parse(values[0])
		if err != nil {
			merr := merry.Errorf("presented Referer %q is not a URL", values[0])
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
			return p.ErrorHandler(c, r, merr)
		}
		if !sameOrigin(expected, actual) {
			merr := merry.Errorf("presented Referer %q invalid, expected: %q", values[0], siteUrl)
			merr = merr.WithHTTPCode(http.StatusInternalServerError)
			return p.ErrorHandler(c, r, merr)
		}
	} else {
		merr := merry.New("no Origin or Referer header presented")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return p.ErrorHandler(c, r, merr)
	}

	cookie, err := r.Cookie(p.CookieName)
	if err != nil {
		merr := merry.Wrap(err)
		merr = merr.WithMessage("no CSRF cookie presented")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return p.ErrorHandler(c, r, merr)
	}
	if len(cookie.Value) != p.TokenLength {
		merr := merry.New("invalid CSRF cookie presented")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return p.ErrorHandler(c, r, merr)
	}

	token, err := p.ExtractToken(r)
	if err != nil {
		merr := merry.Wrap(err)
		merr = merr.WithMessage("invalid CSRF token")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return p.ErrorHandler(c, r, merr)
	}
	if len(token) != p.TokenLength {
		merr := merry.New("invalid CSRF token")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return p.ErrorHandler(c, r, merr)
	}

	if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) != 1 {
		merr := merry.New("invalid CSRF token")
		merr = merr.WithHTTPCode(http.StatusForbidden)
		return p.ErrorHandler(c, r, merr)
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
