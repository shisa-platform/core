package main

import (
	"context"
	"expvar"
	"flag"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	"github.com/percolate/shisa/examples/rpc/service"
	"github.com/percolate/shisa/httpx"
)

const timeFormat = "2006-01-02T15:04:05+00:00"

func main() {
	start := time.Now().UTC()

	startTime := new(expvar.String)
	startTime.Set(start.Format(timeFormat))

	helloVar := expvar.NewMap("hello")
	helloVar.Set("start-time", startTime)
	helloVar.Set("uptime", expvar.Func(func() interface{} {
		now := time.Now().UTC()
		return now.Sub(start).String()
	}))

	addr := flag.String("addr", ":0", "service address")
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("initializing logger: %v", err)
	}

	defer logger.Sync()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	service := &hello.Hello{Logger: logger}
	rpc.Register(service)
	rpc.HandleHTTP()

	server := http.Server{}

	listener, err := httpx.HTTPListenerForAddress(*addr)
	if err != nil {
		logger.Error("opening listener", zap.Error(err))
	}
	logger.Info("starting hello service", zap.String("addr", listener.Addr().String()))

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()

	select {
	case err := <-errCh:
		if !merry.Is(err, http.ErrServerClosed) {
			logger.Error("starting server", zap.Error(err))
		}
	case <-interrupt:
		server.Shutdown(context.Background())
	}
}
