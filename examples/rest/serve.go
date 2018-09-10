package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"
	"go.uber.org/zap"

	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/lb"
	"github.com/shisa-platform/core/sd"
)

func serve(logger *zap.Logger, addr string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		logger.Fatal("consul client failed to initialize", zap.Error(err))
	}
	reg := sd.NewConsul(client)
	bal := lb.NewLeastN(reg, 2)

	service := &Goodbye{
		Balancer: bal,
		Logger:   logger,
	}
	server := http.Server{
		Handler: service,
	}

	listener, err := httpx.HTTPListenerForAddress(addr)
	if err != nil {
		logger.Fatal("opening listener", zap.Error(err))
	}
	logger.Info("starting rest service", zap.String("addr", listener.Addr().String()))

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()

	saddr := listener.Addr().String()
	ru := &url.URL{
		Host:     saddr,
		RawQuery: fmt.Sprintf("id=%s", saddr),
	}
	if err := reg.Register(serviceName, ru); err != nil {
		logger.Fatal("service failed to register", zap.Error(err))
	}
	defer reg.Deregister(serviceName)

	cu := &url.URL{
		Scheme:   "http",
		Host:     saddr,
		Path:     "/healthcheck",
		RawQuery: fmt.Sprintf("interval=5s&id=%s&serviceid=%s", saddr, saddr),
	}

	if err := reg.AddCheck(serviceName, cu); err != nil {
		logger.Fatal("healthcheck failed to register", zap.Error(err))
	}
	defer reg.RemoveChecks(serviceName)

	select {
	case err := <-errCh:
		if !merry.Is(err, http.ErrServerClosed) {
			logger.Error("starting server", zap.Error(err))
		}
	case <-interrupt:
		server.Shutdown(context.Background())
	}
}
