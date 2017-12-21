package authn

import (
	"fmt"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

type genericAuthProvider struct {
	extractor TokenExtractor
	idp       IdentityProvider
	scheme    string
	realm     string
	challenge string
}

func (m *genericAuthProvider) Authenticate(ctx context.Context, r *service.Request) (models.User, merry.Error) {
	credentials, err := m.extractor(ctx, r)
	if err != nil {
		return nil, err
	}

	return m.idp.Authenticate(credentials)
}

func (m *genericAuthProvider) Challenge() string {
	if m.challenge == "" {
		m.challenge = fmt.Sprintf("%s realm=%q", m.scheme, m.realm)
	}

	return m.challenge
}

// NewProvider returns a provider using the given token
// extractor and identity provider.
// An error will be returned if the `idp` or `extractor`
// parameters are nil.
func NewProvider(extractor TokenExtractor, idp IdentityProvider, scheme, realm string) (Provider, merry.Error) {
	if idp == nil {
		return nil, merry.New("Identity provider must be non-nil")
	}
	if extractor == nil {
		return nil, merry.New("Token extractor must be non-nil")
	}

	return &genericAuthProvider{
		extractor: extractor,
		idp:       idp,
		scheme:    scheme,
		realm:     realm,
	}, nil
}
