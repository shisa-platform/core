package server

import (
	"time"
)

type Server interface {
	Name() string
	Addr() string
	Serve() error
	Shutdown(time.Duration) error
}
