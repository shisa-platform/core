package main

import (
	"os"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/auxiliary"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/gateway"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/middleware"
	"github.com/percolate/shisa/service"
)

func serve(logger *zap.Logger, addr, debugAddr, healthcheckAddr string) {
	idp := &ExampleIdentityProvider{Env: env.DefaultProvider}

	authenticator, err := authn.NewBasicAuthenticator(idp, "example")
	if err != nil {
		logger.Fatal("creating authenticator", zap.Error(err))
	}
	authN := &middleware.Authentication{Authenticator: authenticator}

	lh := logHandler{logger}

	gw := &gateway.Gateway{
		Address:         addr,
		HandleInterrupt: true,
		GracePeriod:     2 * time.Second,
		Handlers:        []service.Handler{authN.Service},
		Logger:          logger,
		CompletionHook:  lh.completion,
		ErrorHook:       lh.error,
	}

	authZ := SimpleAuthorization{[]string{"user:1"}}
	debug := &auxiliary.DebugServer{
		HTTPServer: auxiliary.HTTPServer{
			Addr:           debugAddr,
			Authentication: authN,
			Authorizer:     authZ,
			CompletionHook: lh.completion,
			ErrorHook:      lh.error,
		},
		Logger: logger,
	}

	hello := NewHelloService(env.DefaultProvider)
	goodbye := NewGoodbyeService(env.DefaultProvider)

	healthcheck := &auxiliary.HealthcheckServer{
		HTTPServer: auxiliary.HTTPServer{
			Addr:           healthcheckAddr,
			Authentication: authN,
			Authorizer:     authZ,
			CompletionHook: lh.completion,
			ErrorHook:      lh.error,
		},
		Checkers: []auxiliary.Healthchecker{idp, hello, goodbye},
		Logger:   logger,
	}

	services := []service.Service{hello, goodbye}

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

type logHandler struct {
	logger *zap.Logger
}

func (l logHandler) completion(c context.Context, r *service.Request, s httpx.ResponseSnapshot) {
	fs := make([]zapcore.Field, 9, 10+len(s.Metrics))
	fs[0] = zap.String("request-id", c.RequestID())
	fs[1] = zap.String("client-ip-address", r.ClientIP())
	fs[2] = zap.String("method", r.Method)
	fs[3] = zap.String("uri", r.URL.RequestURI())
	fs[4] = zap.Int("status-code", s.StatusCode)
	fs[5] = zap.Int("response-size", s.Size)
	fs[6] = zap.String("user-agent", r.UserAgent())
	fs[7] = zap.Time("start", s.Start)
	fs[8] = zap.Duration("elapsed", s.Elapsed)
	if u := c.Actor(); u != nil {
		fs = append(fs, zap.String("user-id", u.ID()))
	}
	for k, v := range s.Metrics {
		fs = append(fs, zap.Duration(k, v))
	}
	l.logger.Info("request", fs...)
}

func (l logHandler) error(ctx context.Context, _ *service.Request, err merry.Error) {
	l.logger.Error(err.Error(), zap.String("request-id", ctx.RequestID()), zap.Error(err))
}
