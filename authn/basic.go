package authn

import (
	"encoding/base64"
	"fmt"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

// BasicAuthTokenExtractor returns the decoded credentials from
// a Basic Authentication challenge.  The token returned is the
// colon-concatentated userid-password as specified in RFC 7617.
// An error is returned if the credentials cannot be extracted.
func BasicAuthTokenExtractor(ctx context.Context, r *service.Request) (token string, err merry.Error) {
	rawCredentails, err := AuthenticationHeaderTokenExtractor(ctx, r, "Basic")
	if err != nil {
		return
	}

	credentials, b64err := base64.StdEncoding.DecodeString(rawCredentails)
	if b64err != nil {
		err = merry.Wrap(b64err)
		err = err.WithUserMessage("Malformed authentication credentials were provided")
		err = err.WithValue("credentials", rawCredentails)
		return
	}

	token = string(credentials)
	return
}

// BasicAuthnProvider implements Basic Access Authentication as
// specified in RFC 7617.
type BasicAuthnProvider struct {
	IdP       IdentityProvider
	Realm     string
	challenge string
}

func (m *BasicAuthnProvider) Authenticate(ctx context.Context, r *service.Request) (models.User, merry.Error) {
	credentials, err := BasicAuthTokenExtractor(ctx, r)
	if err != nil {
		return nil, err
	}

	return m.IdP.Authenticate(credentials)
}

func (m *BasicAuthnProvider) Challenge() string {
	if m.challenge == "" {
		m.challenge = fmt.Sprintf("Basic realm=%q", m.Realm)
	}

	return m.challenge
}
