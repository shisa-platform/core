package middleware

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

var (
	WWWAuthenticateHeaderKey = http.CanonicalHeaderKey("WWW-Authenticate")
)

// Authentication is middleware to help automate authentication.
//
// `Authenticator` must be non-nil or an InternalServiceError
// status response will be returned.
// `UnauthorizedHandler` can be set to optionally customize the
// response for an unknown user.  The default handler will
// return a 401 status code, the "WWW-Authenticate" header and an
// empty body.
// `ErrorHandler` can be set to optionally customize the response
// for an error. The `err` parameter passed to the handler will
// have a recommended HTTP status code. The default handler will
// return the recommended status code, the "WWW-Authenticate"
// header and an empty body.
type Authentication struct {
	Authenticator       authn.Authenticator
	UnauthorizedHandler service.Handler
	ErrorHandler        service.ErrorHandler
}

func (m *Authentication) Service(ctx context.Context, r *service.Request) service.Response {
	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}
	if m.UnauthorizedHandler == nil {
		m.UnauthorizedHandler = m.defaultHandler
	}

	if m.Authenticator == nil {
		err := merry.New("authn.Authenticator is nil")
		err = err.WithUserMessage("Authenticator.Authenticator must be non-nil")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return m.ErrorHandler(ctx, r, err)
	}

	user, err := m.Authenticator.Authenticate(ctx, r)
	if err != nil {
		err = err.WithHTTPCode(http.StatusUnauthorized)
		return m.ErrorHandler(ctx, r, err)
	}
	if user == nil {
		return m.UnauthorizedHandler(ctx, r)
	}

	ctx = ctx.WithActor(user)
	return nil
}

func (m *Authentication) defaultHandler(ctx context.Context, r *service.Request) service.Response {
	response := service.NewEmpty(http.StatusUnauthorized)
	response.Headers().Set(WWWAuthenticateHeaderKey, m.Authenticator.Challenge())

	return response
}

func (m *Authentication) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	response := service.NewEmpty(merry.HTTPCode(err))
	if m.Authenticator != nil {
		response.Headers().Set(WWWAuthenticateHeaderKey, m.Authenticator.Challenge())
	}

	return response
}

// PassiveAuthentication is middleware to help automate optional
// authentication.
// If the authenticator returns a principal it will be added to
// the context. An error response will never be generated from
// the results of the authenticator.
// `Authenticator` must be non-nil or an InternalServiceError
// status response will be returned.
type PassiveAuthentication struct {
	Authenticator authn.Authenticator
}

func (m *PassiveAuthentication) Service(ctx context.Context, r *service.Request) service.Response {
	if m.Authenticator == nil {
		return service.NewEmpty(http.StatusInternalServerError)
	}

	if user, _ := m.Authenticator.Authenticate(ctx, r); user != nil {
		ctx = ctx.WithActor(user)
	}

	return nil
}
