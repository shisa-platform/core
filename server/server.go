package server

import (
	"time"
)

//go:generate charlatan -output=./server_charlatan.go Server

type Server interface {
	Name() string
	Addr() string
	Serve() error
	Shutdown(time.Duration) error
}
