package models

type User interface {
	Stringer
	ID() string
}
