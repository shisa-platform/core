package main

import (
	"context"
	"expvar"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"
	"go.uber.org/zap"

	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/sd"
)

const timeFormat = "2006-01-02T15:04:05+00:00"

var goodbye = expvar.NewMap("goodbye")

func main() {
	start := time.Now().UTC()

	startTime := new(expvar.String)
	startTime.Set(start.Format(timeFormat))
	goodbye.Set("start-time", startTime)
	goodbye.Set("uptime", expvar.Func(func() interface{} {
		now := time.Now().UTC()
		return now.Sub(start).String()
	}))

	addr := flag.String("addr", "", "service address")
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("initializing logger: %v", err)
	}

	defer logger.Sync()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	service := &Goodbye{Logger: logger}
	server := http.Server{
		Handler: service,
	}

	listener, err := httpx.HTTPListenerForAddress(*addr)
	if err != nil {
		logger.Error("opening listener", zap.Error(err))
	}
	logger.Info("starting goodbye service", zap.String("addr", listener.Addr().String()))

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()

	address, sport, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatalf("parsing addr/port: %v", err)
	}

	port, err := strconv.Atoi(sport)
	if err != nil {
		log.Fatalf("parsing port: %v", err)
	}

	conf := consul.DefaultConfig()
	c, err := consul.NewClient(conf)
	if err != nil {
		panic(err)
	}

	reg := sd.NewConsulRegistrar(c, &consul.AgentServiceRegistration{
		ID:      "goodbye",
		Name:    "goodbye",
		Port:    port,
		Address: address,
	})

	err = reg.Register()
	defer reg.Deregister()

	if err != nil {
		panic(err)
	}
	select {
	case err := <-errCh:
		if !merry.Is(err, http.ErrServerClosed) {
			logger.Error("starting server", zap.Error(err))
		}
	case <-interrupt:
		server.Shutdown(context.Background())
	}
}
