package middleware

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
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
	// the "WWW-Authenticate" header and an empty body.
	ErrorHandler httpx.ErrorHandler
}

func (m *Authentication) Service(ctx context.Context, request *httpx.Request) httpx.Response {
	if m.Authenticator == nil {
		err := merry.New("authentication proxy authenticator is nil")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return m.handleError(ctx, request, err)
	}

	user, err := m.authenticate(ctx, request)
	if err != nil {
		err = err.Prepend("running authentication middleware authenticator")
		err = err.WithHTTPCode(http.StatusUnauthorized)
		return m.handleError(ctx, request, err)
	}
	if user == nil {
		return m.handleUnauthorized(ctx, request)
	}

	ctx = ctx.WithActor(user)

	return nil
}

func (m *Authentication) authenticate(ctx context.Context, request *httpx.Request) (user models.User, err merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic in authenticator")
			return
		}

		err = merry.New("panic in authenticator").WithValue("context", arg)
	}()

	return m.Authenticator.Authenticate(ctx, request)
}

func (m *Authentication) handleUnauthorized(ctx context.Context, request *httpx.Request) httpx.Response {
	if m.UnauthorizedHandler == nil {
		response := httpx.NewEmpty(http.StatusUnauthorized)
		response.Headers().Set(WWWAuthenticateHeaderKey, m.Authenticator.Challenge())
		return response
	}

	var exception merry.Error
	response := m.UnauthorizedHandler.InvokeSafely(ctx, request, &exception)
	if exception != nil {
		exception = exception.Prepend("running authentication middleware unauthorized handler")
		exception = exception.WithHTTPCode(http.StatusUnauthorized)
		response = m.handleError(ctx, request, exception)
	}

	return response
}

func (m *Authentication) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	if m.ErrorHandler == nil {
		response := httpx.NewEmptyError(merry.HTTPCode(err), err)
		if m.Authenticator != nil && merry.HTTPCode(err) == http.StatusUnauthorized {
			challenge := m.Authenticator.Challenge()
			response.Headers().Set(WWWAuthenticateHeaderKey, challenge)
		}

		return response
	}

	var exception merry.Error
	response := m.ErrorHandler.InvokeSafely(ctx, request, err, &exception)
	if exception != nil {
		exception = exception.Prepend("running authentication middleware ErrorHandler")
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
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

		var err merry.Error
		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic in authenticator")
		} else {
			err = merry.New("panic in authenticator").WithValue("context", arg)
		}

		err = err.WithHTTPCode(http.StatusInternalServerError)
		response = httpx.NewEmptyError(http.StatusInternalServerError, err)
	}()

	if m.Authenticator == nil {
		err := merry.New("passive authentication proxy authenticator is nil")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return httpx.NewEmptyError(http.StatusInternalServerError, err)
	}

	if user, _ := m.Authenticator.Authenticate(ctx, request); user != nil {
		ctx = ctx.WithActor(user)
	}

	return nil
}
