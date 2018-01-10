package authn

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

// BasicAuthTokenExtractor returns the decoded credentials from
// a Basic Authentication challenge. The token returned is the
// colon-concatentated userid-password as specified in RFC 7617.
// An error is returned if the credentials cannot be extracted.
func BasicAuthTokenExtractor(ctx context.Context, r *service.Request) (token string, err merry.Error) {
	challenge := strings.TrimSpace(r.Header.Get(AuthnHeaderKey))
	if challenge == "" {
		err = merry.New("no challenge provided")
		err = err.WithUserMessage("Authentication challenge was missing")
		return
	}

	const prefix = "Basic "
  	if !strings.HasPrefix(challenge, prefix) {
		err = merry.New("unsupported authn scheme")
		err = err.WithUserMessage("Unsupported authentication scheme was specified")
  		return
  	}
 
	credentials, b64err := base64.StdEncoding.DecodeString(challenge[len(prefix):])
	if b64err != nil {
		err = merry.Wrap(b64err)
		err = err.WithUserMessage("Malformed authentication credentials were provided")
		return
	}

	token = string(credentials)
	return
}

type basicAuthenticator struct {
	idp       IdentityProvider
	challenge string
}

func (m *basicAuthenticator) Authenticate(ctx context.Context, r *service.Request) (models.User, merry.Error) {
	credentials, err := BasicAuthTokenExtractor(ctx, r)
	if err != nil {
		return nil, err
	}

	return m.idp.Authenticate(credentials)
}

func (m *basicAuthenticator) Challenge() string {
	return m.challenge
}

// NewBasicAuthenticator returns an authenticator implementing Basic
// Access Authentication as specified in RFC 7617.
// An error will be returned if the `idp` parameter is nil.
func NewBasicAuthenticator(idp IdentityProvider, realm string) (Authenticator, merry.Error) {
	if idp == nil {
		return nil, merry.New("Identity provider must be non-nil")
	}

	return &basicAuthenticator{idp: idp, challenge: fmt.Sprintf("Basic realm=%q", realm)}, nil
}
