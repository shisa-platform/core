package authn

import (
	"github.com/ansel1/merry"
	"strings"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
)

// AuthenticationHeaderTokenExtractor returns the token from the
// "Authentication" header.
// An error is returned if the "Authentication" header is missing,
// its value is empty or presented scheme doesn't match the value
// of `scheme`.
func AuthenticationHeaderTokenExtractor(ctx context.Context, r *httpx.Request, scheme string) (token string, err merry.Error) {
	challenge := strings.TrimSpace(r.Header.Get(AuthnHeaderKey))
	if challenge == "" {
		err = merry.New("extract header token: no challenge provided")
		return
	}

	if !strings.HasPrefix(challenge, scheme+" ") {
		err = merry.New("extract header token: unsupported scheme").Append(scheme)
		return
	}

	token = challenge[len(scheme)+1:]
	return
}

// URLTokenExtractor returns the credentials from the request URL
// concatenated together with a colon as specified in RFC 7617.
// An error is returned if the credentials cannot be extracted.
func URLTokenExtractor(ctx context.Context, r *httpx.Request) (token string, err merry.Error) {
	if r.URL.User != nil {
		token = r.URL.User.String()
		return
	}

	err = merry.New("extract url token: no user info found")
	return
}
