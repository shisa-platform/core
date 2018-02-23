package auxiliary

import (
	"time"
)

//go:generate charlatan -output=./server_charlatan.go Server

type Server interface {
	Name() string
	Address() string
	Listen() error
	Serve() error
	Shutdown(time.Duration) error
}
