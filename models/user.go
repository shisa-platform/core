package models

import (
	"fmt"
)

type User interface {
	fmt.Stringer
	ID() string
}
