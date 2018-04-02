package middleware

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/errorx"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/models"
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
	if m.Authenticator == nil {
		err := merry.New("authentication middleware: check invariants: authenticator is nil")
		return m.HandleError(ctx, request, err)
	}

	user, err, exception := m.authenticate(ctx, request)
	if exception != nil {
		exception = exception.Prepend("authentication middleware: authenticate")
		return m.HandleError(ctx, request, exception)
	} else if err != nil {
		err = err.Prepend("authentication middleware: authenticate")
		err = err.WithHTTPCode(http.StatusUnauthorized)
		return m.HandleError(ctx, request, err)
	} else if user == nil {
		return m.HandleUnauthorized(ctx, request)
	}

	ctx = ctx.WithActor(user)

	return nil
}

func (m *Authentication) authenticate(ctx context.Context, request *httpx.Request) (user models.User, err merry.Error, exception merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		exception = errorx.CapturePanic(arg, "panic in authenticator")
	}()

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

func (m *PassiveAuthentication) Service(ctx context.Context, request *httpx.Request) (response httpx.Response) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		err := errorx.CapturePanic(arg, "passive authentication middleware: authenticate: panic in authenticator")
		response = httpx.NewEmptyError(http.StatusInternalServerError, err)
	}()

	if m.Authenticator == nil {
		err := merry.New("passive authentication middleware: check invariants: authenticator nil")
		return httpx.NewEmptyError(http.StatusInternalServerError, err)
	}

	if user, _ := m.Authenticator.Authenticate(ctx, request); user != nil {
		ctx = ctx.WithActor(user)
	}

	return nil
}
