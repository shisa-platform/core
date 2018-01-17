package main

import (
	"expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/auxiliary"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/gateway"
	"github.com/percolate/shisa/middleware"
	"github.com/percolate/shisa/service"
)

const (
	defaultPort            = 9001
	defaultDebugPort       = defaultPort + 1
	defaultHealthcheckPort = defaultDebugPort + 1
	timeFormat             = "2006-01-02T15:04:05+00:00"
)

type authorizer struct{}

func (a authorizer) Authorize(ctx context.Context, r *service.Request) (bool, merry.Error) {
	return ctx.Actor().ID() == "0", nil
}

type dependencyStub struct{}

func (d dependencyStub) Name() string {
	return "stub"
}

func (d dependencyStub) Healthcheck() merry.Error {
	return nil
}

func main() {
	now := time.Now().UTC().Format(timeFormat)
	startTime := expvar.NewString("starttime")
	startTime.Set(now)

	port := flag.Int("port", defaultPort, "Specify service port")
	debugPort := flag.Int("debugport", defaultDebugPort, "Specify debug port")
	healthcheckPort := flag.Int("healthcheckport", defaultHealthcheckPort, "Specify healthcheck port")

	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	defer logger.Sync()

	idp := &SimpleIdentityProvider{
		Users: []User{User{"0", "Admin", "password"}, User{"1", "Boss", "password"}},
	}

	authenticator, err := authn.NewBasicAuthenticator(idp, "example")
	if err != nil {
		panic(err)
	}
	authN := middleware.Authentication{Authenticator: authenticator}
	authZ := authorizer{}

	gw := &gateway.Gateway{
		Name:            "hello",
		Address:         fmt.Sprintf(":%d", *port),
		HandleInterrupt: true,
		GracePeriod:     2 * time.Second,
		Authentication:  &authN,
		Logger:          logger,
	}

	debug := &auxiliary.DebugServer{
		HTTPServer: auxiliary.HTTPServer{
			Addr:           fmt.Sprintf(":%d", *debugPort),
			Authentication: &authN,
			Authorizer:     authZ,
		},
		Logger: logger,
	}
	healthcheck := &auxiliary.HealthcheckServer{
		HTTPServer: auxiliary.HTTPServer{
			Addr:           fmt.Sprintf(":%d", *healthcheckPort),
			Authentication: &authN,
			Authorizer:     authZ,
		},
		Checkers: []auxiliary.Healthchecker{dependencyStub{}},
		Logger:   logger,
	}

	services := []service.Service{NewHelloService(), NewGoodbyeService()}
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
