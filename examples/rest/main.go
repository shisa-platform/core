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
	"github.com/percolate/shisa/auxillary"
	"github.com/percolate/shisa/gateway"
	"github.com/percolate/shisa/middleware"
	"github.com/percolate/shisa/service"
)

const (
	defaultPort      = 9001
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

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	defer logger.Sync()

	idp := &SimpleIdentityProvider{
		Users: []User{User{"1", "Boss", "password"}},
	}
	authenticator, err := authn.NewBasicAuthenticator(idp, "example")
	if err != nil {
		panic(err)
	}
	authN := middleware.Authentication{Authenticator: authenticator}

	gw := &gateway.Gateway{
		Name:            "hello",
		Address:         fmt.Sprintf(":%d", *port),
		HandleInterrupt: true,
		GracePeriod:     2 * time.Second,
		Authentication:  &authN,
		Logger:          logger,
	}

	debug := &auxillary.DebugServer{
		HTTPServer: auxillary.HTTPServer{
			Addr: fmt.Sprintf(":%d", *debugPort),
		},
		Logger: logger,
	}

	services := []service.Service{NewHelloService(), NewGoodbyeService()}
	if err := gw.Serve(services, debug); err != nil {
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
