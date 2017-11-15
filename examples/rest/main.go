package main

import (
	"expvar"
	"flag"
	"fmt"
	"time"

	"github.com/percolate/shisa/gateway"
	"github.com/percolate/shisa/server"
	"go.uber.org/zap"
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

	logger, _ := zap.NewProduction()

	gw := &gateway.Gateway{
		Name:            "hello",
		Address:         fmt.Sprintf(":%d", *port),
		HandleInterrupt: true,
		GracePeriod:     2 * time.Second,
		Logger: logger,
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
