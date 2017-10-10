package modelstest

import (
	"github.com/percolate/shisa/models"
)

type FakeUser struct {
	id string
}

func (u *FakeUser) String() string {
	return u.id
}

func (u *FakeUser) ID() string {
	return u.id
}

func MakeUser(id string) models.User {
	return &FakeUser{id: id}
}
