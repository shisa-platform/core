package authn

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/models"
)

// BasicAuthTokenExtractor returns the decoded credentials from
// a Basic Authentication challenge. The token returned is the
// colon-concatentated userid-password as specified in RFC 7617.
// An error is returned if the credentials cannot be extracted.
func BasicAuthTokenExtractor(ctx context.Context, r *httpx.Request) (token string, err merry.Error) {
	challenge := strings.TrimSpace(r.Header.Get(AuthnHeaderKey))
	if challenge == "" {
		err = merry.New("extract basic auth token: no challenge provided")
		return
	}

	const prefix = "Basic "
	if !strings.HasPrefix(challenge, prefix) {
		err = merry.New("extract basic auth token: unsupported scheme")
		return
	}

	credentials, b64err := base64.StdEncoding.DecodeString(challenge[len(prefix):])
	if b64err != nil {
		err = merry.Prepend(b64err, "extract basic auth token")
		return
	}

	token = string(credentials)
	return
}

type basicAuthenticator struct {
	idp       IdentityProvider
	challenge string
}

func (m *basicAuthenticator) Authenticate(ctx context.Context, r *httpx.Request) (models.User, merry.Error) {
	credentials, err := BasicAuthTokenExtractor(ctx, r)
	if err != nil {
		return nil, err.Prepend("basic auth: authenticate")
	}

	return m.idp.Authenticate(ctx, credentials)
}

func (m *basicAuthenticator) Challenge() string {
	return m.challenge
}

// NewBasicAuthenticator returns an authenticator implementing Basic
// Access Authentication as specified in RFC 7617.
// An error will be returned if the `idp` parameter is nil.
func NewBasicAuthenticator(idp IdentityProvider, realm string) (Authenticator, merry.Error) {
	if idp == nil {
		return nil, merry.New("basic auth: check invariants: identity provider nil")
	}

	return &basicAuthenticator{idp: idp, challenge: fmt.Sprintf("Basic realm=%q", realm)}, nil
}
