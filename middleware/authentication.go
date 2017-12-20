package middleware

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

var (
	wwwAuthenticateHeaderKey = http.CanonicalHeaderKey("WWW-Authenticate")
)

// Authenticator is middleware to help automate authentication.
//
// `Provider` must be non-nil or a InternalServiceError status
// response will be returned.
// `Challenge` should be provided and will be returned as the
// value of the "WWW-Authenticate" response header if the default
// handlers are invoked.
// `UnauthorizedHandler` can be set to optionally customize the
// response for an unknown user.  The default handler will
// return a 401 status code, the "WWW-Authenticate" header and an
// empty body.
// `ErrorHandler` can be set to optionally customize the response
// for an error. The `err` parameter passed to the handler will
// have a recommended HTTP status code. The default handler will
// return the recommended status code, the "WWW-Authenticate"
// header and an empty body.
type Authenticator struct {
	Provider            authn.Provider
	Challenge           string
	UnauthorizedHandler service.Handler
	ErrorHandler        service.ErrorHandler
}

func (m *Authenticator) Service(ctx context.Context, r *service.Request) service.Response {
	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}
	if m.UnauthorizedHandler == nil {
		m.UnauthorizedHandler = m.defaultHandler
	}

	if m.Provider == nil {
		err := merry.New("authn.Provider is nil")
		err = err.WithUserMessage("Authenticator.Provider must be non-nil")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return m.ErrorHandler(ctx, r, err)
	}

	user, err := m.Provider.Authenticate(ctx, r)
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

func (m *Authenticator) defaultHandler(ctx context.Context, r *service.Request) service.Response {
	response := service.NewEmpty(http.StatusUnauthorized)
	response.Headers().Set(wwwAuthenticateHeaderKey, m.Challenge)

	return response
}

func (m *Authenticator) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	response := service.NewEmpty(merry.HTTPCode(err))
	response.Headers().Set(wwwAuthenticateHeaderKey, m.Challenge)

	return response
}