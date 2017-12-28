package authn

import (
	"fmt"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

type bearerAuthenticator struct {
	idp       IdentityProvider
	realm     string
	challenge string
}

func (m *bearerAuthenticator) Authenticate(ctx context.Context, r *service.Request) (models.User, merry.Error) {
	credentials, err := AuthenticationHeaderTokenExtractor(ctx, r, "Bearer")
	if err != nil {
		return nil, err
	}

	return m.idp.Authenticate(credentials)
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
		return nil, merry.New("Identity provider must be non-nil")
	}

	return &bearerAuthenticator{idp: idp, realm: realm}, nil
}
