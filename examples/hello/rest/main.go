package main

import (
	"fmt"
	"expvar"
	"flag"
	"time"

	"github.com/percolate/shisa/gateway"
	"github.com/percolate/shisa/server"
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
		Name: "hello",
		Address: fmt.Sprintf(":%d", *port),
		HandleInterrupt: true,
		GracePeriod: 2 * time.Second,
	}
	debug := &server.DebugServer{
		Address: fmt.Sprintf(":%d", *debugPort),
	}

	gw.RegisterAuxillary(debug)

	err := gw.Serve()
	if err != nil {
		fmt.Printf("uh oh! %v", err)
	}
}
