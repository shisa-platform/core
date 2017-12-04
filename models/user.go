package models

import (
	"fmt"
)

//go:generate charlatan -output=./user_charlatan.go User

type User interface {
	fmt.Stringer
	ID() string
}
