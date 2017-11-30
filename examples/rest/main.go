package main

import (
	"expvar"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/percolate/shisa/gateway"
	"github.com/percolate/shisa/server"
	"github.com/percolate/shisa/service"
)

const (
	defaultPort      = 8000
	defaultDebugPort = defaultPort + 1
	timeFormat       = "2006-01-02T15:04:05+00:00"
)

func main() {
	now := time.Now().UTC().Format(timeFormat)
	startTime := expvar.NewString("starttime")
	startTime.Set(now)

	port := flag.Int("port", defaultPort, "Specify service port")
	debugPort := flag.Int("debugport", defaultDebugPort, "Specify debug port")

	flag.Parse()

	gw := &gateway.Gateway{
		Name:            "hello",
		Address:         fmt.Sprintf(":%d", *port),
		HandleInterrupt: true,
		GracePeriod:     2 * time.Second,
	}
	debug := &server.DebugServer{
		Address: fmt.Sprintf(":%d", *debugPort),
	}

	services := []Service{NewService()}
	if err := gw.Serve(services, debug); err != nil {
		fmt.Printf("uh oh! %v", err)
	}
}
