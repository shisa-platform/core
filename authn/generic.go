package authn

import (
	"fmt"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/models"
)

type genericAuthenticator struct {
	extractor httpx.StringExtractor
	idp       IdentityProvider
	scheme    string
	realm     string
	challenge string
}

func (m *genericAuthenticator) Authenticate(ctx context.Context, r *httpx.Request) (models.User, merry.Error) {
	credentials, err := m.extractor(ctx, r)
	if err != nil {
		return nil, err.Prepend("auth: authenticate")
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
func NewAuthenticator(extractor httpx.StringExtractor, idp IdentityProvider, scheme, realm string) (Authenticator, merry.Error) {
	if idp == nil {
		return nil, merry.New("auth: check invariants: identity provider nil")
	}
	if extractor == nil {
		return nil, merry.New("auth: check invariants: token extractor nil")
	}

	return &genericAuthenticator{
		extractor: extractor,
		idp:       idp,
		scheme:    scheme,
		realm:     realm,
	}, nil
}
