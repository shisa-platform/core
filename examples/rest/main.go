package main

import (
	"expvar"
	"flag"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/percolate/shisa/auxillary"
	"github.com/percolate/shisa/gateway"
	"github.com/percolate/shisa/service"
)

const (
	defaultPort      = 8000
	defaultDebugPort = defaultPort + 1
	timeFormat       = "2006-01-02T15:04:05+00:00"
)

var (
	commonPolicy = service.Policy{
		TimeBudget:                  time.Millisecond * 5,
		AllowTrailingSlashRedirects: true,
	}
)

func main() {
	now := time.Now().UTC().Format(timeFormat)
	startTime := expvar.NewString("starttime")
	startTime.Set(now)

	port := flag.Int("port", defaultPort, "Specify service port")
	debugPort := flag.Int("debugport", defaultDebugPort, "Specify debug port")

	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	defer logger.Sync()

	gw := &gateway.Gateway{
		Name:            "hello",
		Address:         fmt.Sprintf(":%d", *port),
		HandleInterrupt: true,
		GracePeriod:     2 * time.Second,
		Logger:          logger,
	}

	debug := &auxillary.DebugServer{
		HTTPServer: auxillary.HTTPServer{
			Addr: fmt.Sprintf(":%d", *debugPort),
		},
		Logger: logger,
	}

	services := []service.Service{&HelloService{}, &GoodbyeService{}}
	if err := gw.Serve(services, debug); err != nil {
		logger.Fatal("gateway error", zap.Error(err))
	}
}
