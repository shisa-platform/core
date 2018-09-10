package middleware

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
)

const (
	defaultCookieName  = "csrftoken"
	defaultTokenLength = 32
)

// CheckOrigin compares two URLs and determines if they should be
// considered the "same" for the purposes of CSRF protection.
type CheckOrigin func(expected, actual *url.URL) bool

func (h CheckOrigin) InvokeSafely(expected, actual *url.URL) (ok bool, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in check origin hook")

	return h(expected, actual), nil
}

// CSRFProtector is middleware used to guard against CSRF attacks.
type CSRFProtector struct {
	// SiteURL is the URL to use for CSRF protection. This must
	// be non-nil and contain non-empty Scheme and Host values
	// or a internal server error will be returned.
	SiteURL *url.URL

	// IsExempt optionally customizes checking request exemption
	// from CSRF protection.
	// The default checker always returns `false`.
	IsExempt httpx.RequestPredicate

	// CheckOrigin optionally customizes how URLs should be
	// compared for the purposes of CSRF protection.
	// The default comparisons ensures that URL Schemes and Hosts
	// are equal.
	CheckOrigin CheckOrigin

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

	// ErrorHandler optionally customizes the response for an
	// error. The `err` parameter passed to the handler will
	// have a recommended HTTP status code.
	// The default handler will return the recommended status
	// code and an empty body.
	ErrorHandler httpx.ErrorHandler
}

func (m *CSRFProtector) Service(ctx context.Context, request *httpx.Request) httpx.Response {
	subCtx := ctx
	if ctx.Span() != nil {
		var span opentracing.Span
		span, subCtx = context.StartSpan(ctx, "CSRFProtect")
		defer span.Finish()
		ext.Component.Set(span, "middleware")
	}

	if m.SiteURL == nil {
		err := merry.New("csrf middleware: check invariants: SiteURL is nil")
		return m.handleError(subCtx, request, err)
	} else if m.SiteURL.Host == "" {
		err := merry.New("csrf middleware: check invariants: SiteURL host is empty")
		return m.handleError(subCtx, request, err)
	} else if m.SiteURL.Scheme == "" {
		err := merry.New("csrf middleware: check invariants: SiteURL scheme is empty")
		return m.handleError(subCtx, request, err)
	}

	if ok, exception := m.isExempt(subCtx, request); exception != nil {
		return m.handleError(subCtx, request, exception)
	} else if ok {
		return nil
	}

	values, exists := request.Header["Origin"]
	if !exists {
		values, exists = request.Header["Referer"]
	}
	if !exists {
		err := merry.New("csrf middleware: find header: missing Origin or Referer")
		return m.handleError(subCtx, request, err.WithHTTPCode(http.StatusForbidden))
	}
	if len(values) != 1 {
		err1 := merry.New("csrf middleware: validate header: too many values")
		err1 = err1.WithValue("header", strings.Join(values, ", "))
		return m.handleError(subCtx, request, err1.WithHTTPCode(http.StatusForbidden))
	}

	actual, err := url.Parse(values[0])
	if err != nil {
		err1 := merry.Prepend(err, "csrf middleware: parse Origin/Referer URL")
		err1 = err1.WithValue("url", values[0])
		return m.handleError(subCtx, request, err1.WithHTTPCode(http.StatusForbidden))
	}

	if ok, exception := m.checkOrigin(m.SiteURL, actual); exception != nil {
		exception = exception.Prepend("csrf middleware: check origin")
		return m.handleError(subCtx, request, exception)
	} else if !ok {
		err1 := merry.New("csrf middleware: check origin: not equal")
		err1 = err1.WithValue("expected", m.SiteURL.String())
		err1 = err1.WithValue("actual", actual.String())
		return m.handleError(subCtx, request, err1.WithHTTPCode(http.StatusForbidden))
	}

	name := defaultCookieName
	if m.CookieName != "" {
		name = m.CookieName
	}
	cookie, err := request.Cookie(name)
	if err != nil {
		err1 := merry.New("csrf middleware: find cookie: not found")
		err1 = err1.WithValue("name", name)
		return m.handleError(subCtx, request, err1.WithHTTPCode(http.StatusForbidden))
	}

	length := defaultTokenLength
	if m.TokenLength != 0 {
		length = m.TokenLength
	}
	if len(cookie.Value) != length {
		err1 := merry.New("csrf middleware: vaidate cookie: invalid length")
		err1 = err1.WithValue("cookie", cookie.Value)
		return m.handleError(subCtx, request, err1.WithHTTPCode(http.StatusForbidden))
	}

	token, err := m.extractToken(subCtx, request)
	if err != nil {
		err1 := merry.Prepend(err, "csrf middleware: extract token")
		return m.handleError(subCtx, request, err1)
	}

	if len(token) != length {
		err1 := merry.New("csrf middleware: vaidate token: invalid length")
		err1 = err1.WithValue("token", token)
		return m.handleError(subCtx, request, err1.WithHTTPCode(http.StatusForbidden))
	}

	if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) != 1 {
		err1 := merry.New("csrf middleware: compare tokens: not equal")
		err1 = err1.WithValue("cookie", cookie.Value)
		err1 = err1.WithValue("token", token)
		return m.handleError(subCtx, request, err1.WithHTTPCode(http.StatusForbidden))
	}

	return nil
}

func (m *CSRFProtector) isExempt(ctx context.Context, request *httpx.Request) (ok bool, exception merry.Error) {
	if m.IsExempt == nil {
		return false, nil
	}

	ok, exception = m.IsExempt.InvokeSafely(ctx, request)
	if exception != nil {
		exception = exception.Prepend("csrf middleware: run IsExempt")
	}

	return
}

func (m *CSRFProtector) checkOrigin(expected, actual *url.URL) (bool, merry.Error) {
	if m.CheckOrigin == nil {
		return expected.Scheme == actual.Scheme && expected.Host == actual.Host, nil
	}

	return m.CheckOrigin.InvokeSafely(expected, actual)
}

func (m *CSRFProtector) extractToken(ctx context.Context, request *httpx.Request) (token string, err merry.Error) {
	if m.ExtractToken == nil {
		token = request.Header.Get("X-Csrf-Token")
		if token == "" {
			err = merry.New("missing X-Csrf-Token header")
			err = err.WithHTTPCode(http.StatusForbidden)
		}

		return
	}

	var exception merry.Error
	token, err, exception = m.ExtractToken.InvokeSafely(ctx, request)
	if exception != nil {
		err = exception
	} else if err != nil {
		err = err.WithHTTPCode(http.StatusForbidden)
	}

	return
}

func (m *CSRFProtector) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	span := noopSpan
	if ctxSpan := ctx.Span(); ctxSpan != nil {
		span = ctxSpan
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
	}

	if m.ErrorHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err)
	}

	response, exception := m.ErrorHandler.InvokeSafely(ctx, request, err)
	if exception != nil {
		exception = exception.Prepend("csrf middleware: run ErrorHandler")
		span.LogFields(otlog.String("exception", exception.Error()))
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
	}

	return response
}
