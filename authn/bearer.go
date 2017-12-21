package authn

import (
	"fmt"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

type bearerAuthProvider struct {
	idp       IdentityProvider
	realm     string
	challenge string
}

func (m *bearerAuthProvider) Authenticate(ctx context.Context, r *service.Request) (models.User, merry.Error) {
	credentials, err := AuthenticationHeaderTokenExtractor(ctx, r, "Bearer")
	if err != nil {
		return nil, err
	}

	return m.idp.Authenticate(credentials)
}

func (m *bearerAuthProvider) Challenge() string {
	if m.challenge == "" {
		m.challenge = fmt.Sprintf("Bearer realm=%q", m.realm)
	}

	return m.challenge
}

// NewBearerAuthenticationProvider returns a provider
// implementing Bearer Access Authentication as specified in RFC
// 7617.
// An error will be returned if the `idp` parameter is nil.
func NewBearerAuthenticationProvider(idp IdentityProvider, realm string) (Provider, merry.Error) {
	if idp == nil {
		return nil, merry.New("Identity provider must be non-nil")
	}

	return &bearerAuthProvider{idp: idp, realm: realm}, nil
}
