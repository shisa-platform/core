package main

import (
	"strings"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/models"
)

type User struct {
	Ident string
	Name  string
	Pass  string
}

func (u User) ID() string {
	return u.Ident
}

func (u User) String() string {
	return u.Name
}

type SimpleIdentityProvider struct {
	Users []User
}

func (p *SimpleIdentityProvider) Authenticate(credentials string) (models.User, merry.Error) {
	credentialParts := strings.Split(credentials, ":")
	if len(credentialParts) != 2 {
		err := merry.New("wrong credential parts count")
		err = err.WithUserMessage("Malformed Basic Authentication credentials were provided")
		return nil, err
	}

	for _, u := range p.Users {
		if u.Name == credentialParts[0] && u.Pass == credentialParts[1] {
			return u, nil
		}
	}

	return nil, nil
}
