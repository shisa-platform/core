package main

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

var (
	authHeaderKey = http.CanonicalHeaderKey("Authorization")
)

type User struct {
	Ident string
	Name string
	Pass string
}

func (u User) ID() string {
	return u.Ident
}

func (u User) String() string {
	return u.Name
}

type BasicAuthnProvider struct {
	Users []User
}

func (m *BasicAuthnProvider) Authenticate(ctx context.Context, r *service.Request) (models.User, merry.Error) {
	challenge := r.Header.Get(authHeaderKey)
	if challenge == "" {
		err := merry.New("no credentials provided")
		err = err.WithUserMessage("Basic Authentication credentials must be provided")
		return nil, err
	}

	challengeParts := strings.Split(challenge, " ")
	if len(challengeParts) != 2 {
		err := merry.New("too many credential parts")
		err = err.WithUserMessage("Malformed Basic Authentication credentials were provided")
		return nil, err
	}

	if challengeParts[0] != "Basic" {
		err := merry.New("unsupported authn scheme")
		err = err.WithUserMessage("Basic Authentication scheme must be specified")
		return nil, err
	}

	credentials, err := base64.StdEncoding.DecodeString(challengeParts[1])
	if err != nil {
		me := merry.Wrap(err)
		me = me.WithUserMessage("Malformed Basic Authentication credentials were provided")
		return nil, me
	}

	credentialParts := strings.Split(string(credentials), ":")
	if len(credentialParts) != 2 {
		err := merry.New("too many credential parts")
		err = err.WithUserMessage("Malformed Basic Authentication credentials were provided")
		return nil, err
	}

	for _, u := range m.Users {
		if u.Name == credentialParts[0] && u.Pass == credentialParts[1] {
			return u, nil
		}
	}

	return nil, nil
}
