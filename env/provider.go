package env

import (
	"github.com/ansel1/merry"
)

var (
	NameNotSet = merry.New("name not set")
	NameEmpty  = merry.New("name has empty value")
)

type Value struct {
	Name  string
	Value string
}

//go:generate charlatan -output=./provider_charlatan.go Provider

type Provider interface {
	Get(string) (string, merry.Error)
	GetInt(string) (int, merry.Error)
	GetBool(string) (bool, merry.Error)
	Monitor(string, chan<- Value)
	Healthcheck() merry.Error
}
