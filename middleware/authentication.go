package middleware

import (
	"net/http"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/shisa-platform/core/authn"
	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/models"
)

const (
	WWWAuthenticateHeaderKey = "Www-Authenticate"
)

// Authentication is middleware to help automate authentication.
type Authentication struct {
	// Authenticator must be non-nil or an InternalServiceError
	// status response will be returned.
	Authenticator authn.Authenticator

	// UnauthorizedHandler can be set to optionally customize the
	// response for an unknown user.  The default handler will
	// return a 401 status code, the "WWW-Authenticate" header
	// and an empty body.
	UnauthorizedHandler httpx.Handler

	// `ErrorHandler` can be set to optionally customize the
	// response for an error. The `err` parameter passed to the
	// handler will have a recommended HTTP status code. The
	// default handler will return the recommended status code,
	// the "WWW-Authenticate" header (if the recommended status
	// code is 401) and an empty body.
	ErrorHandler httpx.ErrorHandler
}

func (m *Authentication) Service(ctx context.Context, request *httpx.Request) httpx.Response {
	subCtx := ctx
	if ctx.Span() != nil {
		var span opentracing.Span
		span, subCtx = context.StartSpan(ctx, "Authenticate")
		defer span.Finish()
		ext.Component.Set(span, "middleware")
	}

	if m.Authenticator == nil {
		err := merry.New("authentication middleware: check invariants: authenticator is nil")
		return m.HandleError(subCtx, request, err)
	}

	user, err, exception := m.authenticate(subCtx, request)
	if exception != nil {
		exception = exception.Prepend("authentication middleware: authenticate")
		return m.HandleError(subCtx, request, exception)
	} else if err != nil {
		err = err.Prepend("authentication middleware: authenticate")
		err = err.WithHTTPCode(http.StatusUnauthorized)
		return m.HandleError(subCtx, request, err)
	} else if user == nil {
		return m.HandleUnauthorized(subCtx, request)
	}

	ctx = ctx.WithActor(user)

	return nil
}

func (m *Authentication) authenticate(ctx context.Context, request *httpx.Request) (user models.User, err merry.Error, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in authenticator")

	user, err = m.Authenticator.Authenticate(ctx, request)

	return
}

func (m *Authentication) HandleUnauthorized(ctx context.Context, request *httpx.Request) httpx.Response {
	if m.UnauthorizedHandler == nil {
		response := httpx.NewEmpty(http.StatusUnauthorized)
		response.Headers().Set(WWWAuthenticateHeaderKey, m.Authenticator.Challenge())
		return response
	}

	response, exception := m.UnauthorizedHandler.InvokeSafely(ctx, request)
	if exception != nil {
		exception = exception.Prepend("authentication middleware: run UnauthorizedHandler")
		response = m.HandleError(ctx, request, exception)
	}

	if response.StatusCode() == http.StatusUnauthorized && response.Headers().Get(WWWAuthenticateHeaderKey) == "" {
		response.Headers().Set(WWWAuthenticateHeaderKey, m.Authenticator.Challenge())
	}

	return response
}

func (m *Authentication) HandleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	span := noopSpan
	if ctxSpan := ctx.Span(); ctxSpan != nil {
		span = ctxSpan
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
	}

	if m.ErrorHandler == nil {
		response := httpx.NewEmptyError(merry.HTTPCode(err), err)
		if m.Authenticator != nil && merry.HTTPCode(err) == http.StatusUnauthorized {
			response.Headers().Set(WWWAuthenticateHeaderKey, m.Authenticator.Challenge())
		}

		return response
	}

	response, exception := m.ErrorHandler.InvokeSafely(ctx, request, err)
	if exception != nil {
		exception = exception.Prepend("authentication middleware: run ErrorHandler")
		span.LogFields(otlog.String("exception", exception.Error()))
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
	}

	if m.Authenticator != nil && merry.HTTPCode(err) == http.StatusUnauthorized && response.Headers().Get(WWWAuthenticateHeaderKey) == "" {
		response.Headers().Set(WWWAuthenticateHeaderKey, m.Authenticator.Challenge())
	}

	return response
}

// PassiveAuthentication is middleware to help automate optional
// authentication.
// If the authenticator returns a principal it will be added to
// the context. An error response will never be generated if
// no principal is found.
// `Authenticator` must be non-nil or an InternalServiceError
// status response will be returned.  If the Authenticator panics
// an Unauthorized status response will be returned.
type PassiveAuthentication struct {
	Authenticator authn.Authenticator
}

func (m *PassiveAuthentication) Service(ctx context.Context, request *httpx.Request) httpx.Response {
	subCtx := ctx
	span := noopSpan
	if ctx.Span() != nil {
		span, subCtx = context.StartSpan(ctx, "PassiveAuthenticate")
		defer span.Finish()
		ext.Component.Set(span, "middleware")
	}

	if m.Authenticator == nil {
		err := merry.New("passive authentication middleware: check invariants: authenticator nil")
		span.LogFields(otlog.String("error", err.Error()))
		return httpx.NewEmptyError(http.StatusInternalServerError, err)
	}

	if user, exception := m.authenticate(subCtx, request); exception != nil {
		span.LogFields(otlog.String("exception", exception.Error()))
		return httpx.NewEmptyError(http.StatusInternalServerError, exception)
	} else {
		ctx = ctx.WithActor(user)
	}

	return nil
}

func (m *PassiveAuthentication) authenticate(ctx context.Context, request *httpx.Request) (user models.User, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "passive authentication middleware: authenticate: panic in authenticator")

	user, _ = m.Authenticator.Authenticate(ctx, request)

	return
}
