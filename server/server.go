package server

import (
	"time"
)

//go:generate charlatan Server

type Server interface {
	Name() string
	Addr() string
	Serve() error
	Shutdown(time.Duration) error
}
