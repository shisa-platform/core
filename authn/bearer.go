package authn

import (
	"fmt"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/models"
)

type bearerAuthenticator struct {
	idp       IdentityProvider
	realm     string
	challenge string
}

func (m *bearerAuthenticator) Authenticate(ctx context.Context, r *httpx.Request) (models.User, merry.Error) {
	credentials, err := AuthenticationHeaderTokenExtractor(ctx, r, "Bearer")
	if err != nil {
		return nil, err.Prepend("bearer auth: authenticate")
	}

	return m.idp.Authenticate(ctx, credentials)
}

func (m *bearerAuthenticator) Challenge() string {
	if m.challenge == "" {
		m.challenge = fmt.Sprintf("Bearer realm=%q", m.realm)
	}

	return m.challenge
}

// NewBearerAuthenticatior returns an authenticator implementing Bearer
// Access Authentication as specified in RFC 7617.
// An error will be returned if the `idp` parameter is nil.
func NewBearerAuthenticator(idp IdentityProvider, realm string) (Authenticator, merry.Error) {
	if idp == nil {
		return nil, merry.New("bearer auth: check invariants: identity provider nil")
	}

	return &bearerAuthenticator{idp: idp, realm: realm}, nil
}
