package authn

import (
	"github.com/ansel1/merry"
	"strings"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func AuthenticationHeaderTokenExtractor(ctx context.Context, r *service.Request, scheme string) (token string, err merry.Error) {
	challenge := r.Header.Get(authHeaderKey)
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
