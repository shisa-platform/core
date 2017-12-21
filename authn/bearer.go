package authn

import (
	"fmt"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

// BearerAuthnProvider implements Bearer Access Authentication as
// specified in RFC 7617.
type BearerAuthnProvider struct {
	IdP       IdentityProvider
	Realm     string
	challenge string
}

func (m *BearerAuthnProvider) Authenticate(ctx context.Context, r *service.Request) (models.User, merry.Error) {
	credentials, err := AuthenticationHeaderTokenExtractor(ctx, r, "Bearer")
	if err != nil {
		return nil, err
	}

	return m.IdP.Authenticate(credentials)
}

func (m *BearerAuthnProvider) Challenge() string {
	if m.challenge == "" {
		m.challenge = fmt.Sprintf("Bearer realm=%q", m.Realm)
	}

	return m.challenge
}
