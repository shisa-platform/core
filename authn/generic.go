package authn

import (
	"fmt"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

type genericAuthenticator struct {
	extractor service.StringExtractor
	idp       IdentityProvider
	scheme    string
	realm     string
	challenge string
}

func (m *genericAuthenticator) Authenticate(ctx context.Context, r *service.Request) (models.User, merry.Error) {
	credentials, err := m.extractor(ctx, r)
	if err != nil {
		return nil, err
	}

	return m.idp.Authenticate(ctx, credentials)
}

func (m *genericAuthenticator) Challenge() string {
	if m.challenge == "" {
		m.challenge = fmt.Sprintf("%s realm=%q", m.scheme, m.realm)
	}

	return m.challenge
}

// NewAuthenticator returns an authenticator using the given token
// extractor and identity provider.
// An error will be returned if the `idp` or `extractor`
// parameters are nil.
func NewAuthenticator(extractor service.StringExtractor, idp IdentityProvider, scheme, realm string) (Authenticator, merry.Error) {
	if idp == nil {
		return nil, merry.New("Identity provider must be non-nil")
	}
	if extractor == nil {
		return nil, merry.New("Token extractor must be non-nil")
	}

	return &genericAuthenticator{
		extractor: extractor,
		idp:       idp,
		scheme:    scheme,
		realm:     realm,
	}, nil
}
