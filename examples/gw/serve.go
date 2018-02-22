package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/auxiliary"
	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/gateway"
	"github.com/percolate/shisa/lb"
	"github.com/percolate/shisa/middleware"
	"github.com/percolate/shisa/sd"
	"github.com/percolate/shisa/service"
)

const name = "gateway"

func serve(logger *zap.Logger, addr, debugAddr, healthcheckAddr string) {
	conf := consul.DefaultConfig()
	c, e := consul.NewClient(conf)
	if e != nil {
		logger.Fatal("consul failed to initialize", zap.Error(e))
	}

	b := lb.NewRoundRobin()
	res := sd.NewConsul(c, b)

	idp := &ExampleIdentityProvider{Env: env.DefaultProvider, Resolver: res}

	authenticator, err := authn.NewBasicAuthenticator(idp, "example")
	if err != nil {
		logger.Fatal("creating authenticator", zap.Error(err))
	}
	authN := &middleware.Authentication{Authenticator: authenticator}

	gw := &gateway.Gateway{
		Name:            "example",
		Address:         addr,
		HandleInterrupt: true,
		GracePeriod:     2 * time.Second,
		Handlers:        []service.Handler{authN.Service},
		Logger:          logger,
		Registrar:       res,
	}

	authZ := SimpleAuthorization{[]string{"user:1"}}
	debug := &auxiliary.DebugServer{
		HTTPServer: auxiliary.HTTPServer{
			Addr:           debugAddr,
			Authentication: authN,
			Authorizer:     authZ,
		},
		Logger: logger,
	}

	hello := NewHelloService(env.DefaultProvider, res)

	goodbye := NewGoodbyeService(env.DefaultProvider, res)

	healthcheck := &auxiliary.HealthcheckServer{
		HTTPServer: auxiliary.HTTPServer{
			Addr:           healthcheckAddr,
			Authentication: authN,
			Authorizer:     authZ,
		},
		Checkers:  []auxiliary.Healthchecker{idp, hello, goodbye},
		Logger:    logger,
		Registrar: res,
	}

	services := []service.Service{hello, goodbye}

	defer res.Deregister(name)
	defer res.RemoveChecks(name)

	if err := gw.Serve(services, debug, healthcheck); err != nil {
		for _, e := range multierr.Errors(err) {
			values := merry.Values(e)
			fs := make([]zapcore.Field, 0, len(values))
			for name, value := range values {
				if key, ok := name.(string); ok {
					fs = append(fs, zap.Reflect(key, value))
				}
			}
			logger.Error(merry.Message(e), fs...)
		}
		os.Exit(1)
	}
}
