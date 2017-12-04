package service

//go:generate charlatan -output=./service_charlatan.go Service

type Service interface {
	Name() string
	Endpoints() []Endpoint
}
