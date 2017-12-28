package authn

import (
	"github.com/ansel1/merry"
	"strings"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

// AuthenticationHeaderTokenExtractor returns the token from the
// "Authentication" header.
// An error is returned if the "Authentication" header is missing,
// its value is empty or presented scheme doesn't match the value
// of `scheme`.
func AuthenticationHeaderTokenExtractor(ctx context.Context, r *service.Request, scheme string) (token string, err merry.Error) {
	challenge := strings.TrimSpace(r.Header.Get(authHeaderKey))
	if challenge == "" {
		err = merry.New("no challenge provided")
		err = err.WithUserMessage("Authentication challenge was missing")
		return
	}

	challengeParts := strings.Split(challenge, " ")
	if len(challengeParts) != 2 {
		err = merry.New("wrong challenge parts count")
		err = err.WithUserMessage("Malformed authentication challenge was provided")
		err = err.WithValue("challenge", challenge)
		return
	}

	if challengeParts[0] != scheme {
		err = merry.New("unsupported authn scheme")
		err = err.WithUserMessage("Unsupported authentication scheme was specified")
		err = err.WithValue("scheme", challengeParts[0])
		return
	}

	token = challengeParts[1]
	return
}

// URLTokenExtractor returns the credentials from the request URL
// concatenated together with a colon as specified in RFC 7617.
// An error is returned if the credentials cannot be extracted.
func URLTokenExtractor(ctx context.Context, r *service.Request) (token string, err merry.Error) {
	if r.URL.User != nil {
		token = r.URL.User.String()
		return
	}

	err = merry.New("no user info found")
	err = err.WithUserMessage("URL User Info was missing")
	return
}
