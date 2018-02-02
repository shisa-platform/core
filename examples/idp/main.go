package main

import (
	"context"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	service "github.com/percolate/shisa/examples/idp/server"
)

const (
	defaultPort = 9401
	timeFormat  = "2006-01-02T15:04:05+00:00"
)

func main() {
	now := time.Now().UTC().Format(timeFormat)

	startTime := new(expvar.String)
	startTime.Set(now)

	idp := expvar.NewMap("idp")
	idp.Set("start-time", startTime)

	hits := new(expvar.Map)
	idp.Set("hits", hits)

	port := flag.Int("port", defaultPort, "service port")

	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	defer logger.Sync()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	service := &service.Idp{Logger: logger, Hits: hits}
	rpc.Register(service)
	rpc.HandleHTTP()

	server := http.Server{Addr: fmt.Sprintf(":%d", *port)}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting idp server", zap.String("addr", server.Addr))
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !merry.Is(err, http.ErrServerClosed) {
			logger.Error("server failed to start", zap.Error(err))
		}
	case <-interrupt:
		server.Shutdown(context.Background())
	}
}
