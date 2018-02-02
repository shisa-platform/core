package main

import (
	"context"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"
)

const (
	defaultPort = 9601
	timeFormat  = "2006-01-02T15:04:05+00:00"
)

var goodbye = expvar.NewMap("goodbye")

func main() {
	now := time.Now().UTC().Format(timeFormat)
	startTime := new(expvar.String)
	startTime.Set(now)
	goodbye.Set("start-time", startTime)

	port := flag.Int("port", defaultPort, "service port")

	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	defer logger.Sync()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	service := &Goodbye{Logger: logger}
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: service,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting goodbye server", zap.String("addr", server.Addr))
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
