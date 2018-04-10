package main

import (
	"net/url"
	"time"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/auxiliary"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/gateway"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/lb"
	"github.com/percolate/shisa/middleware"
	"github.com/percolate/shisa/sd"
)

func serve(logger *zap.Logger, addr, debugAddr, healthcheckAddr string) {
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		logger.Fatal("consul failed to initialize", zap.Error(err))
	}

	res := sd.NewConsul(client)
	bal := lb.NewLeastN(res, 2)

	idp := &ExampleIdentityProvider{Balancer: bal}

	authenticator, err := authn.NewBasicAuthenticator(idp, "example")
	if err != nil {
		logger.Fatal("creating authenticator", zap.Error(err))
	}
	authN := &middleware.Authentication{Authenticator: authenticator}

	lh := logHandler{logger}

	authZ := SimpleAuthorization{[]string{"user:1"}}
	debug := &auxiliary.DebugServer{
		HTTPServer: auxiliary.HTTPServer{
			Addr:           debugAddr,
			Authentication: authN,
			Authorizer:     authZ,
			CompletionHook: lh.completion,
			ErrorHook:      lh.error,
		},
	}
	go runAuxiliary(debug, logger)

	hello := NewHelloService(bal)
	goodbye := NewGoodbyeService(bal)

	healthcheck := &auxiliary.HealthcheckServer{
		HTTPServer: auxiliary.HTTPServer{
			Addr:           healthcheckAddr,
			Authentication: authN,
			Authorizer:     authZ,
			CompletionHook: lh.completion,
			ErrorHook:      lh.error,
		},
		Checkers: []auxiliary.Healthchecker{idp, hello, goodbye},
	}
	go runAuxiliary(healthcheck, logger)

	gw := &gateway.Gateway{
		Name:            serviceName,
		Addr:            addr,
		HandleInterrupt: true,
		GracePeriod:     2 * time.Second,
		Handlers:        []httpx.Handler{authN.Service},
		Registrar:       res,
		CompletionHook:  lh.completion,
		ErrorHook:       lh.error,
	}

	gw.RegistrationURLHook = func() (u *url.URL, err merry.Error) {
		u = &url.URL{
			Host:     gw.Address(),
			RawQuery: "id=" + gw.Name,
		}
		return
	}
	gw.CheckURLHook = func() (u *url.URL, err merry.Error) {
		u = &url.URL{
			Scheme:   "http",
			Host:     healthcheck.Address(),
			Path:     healthcheck.Path,
			User:     url.UserPassword("Admin", "password"),
			RawQuery: "interval=30s&serviceid=" + gw.Name,
		}
		return
	}

	ch := make(chan merry.Error, 1)
	go func() {
		ch <- gw.Serve(&hello.Service, &goodbye.Service)
	}()

	logger.Info("gateway started", zap.String("addr", gw.Address()))

	select {
	case gwErr := <-ch:
		if gwErr == nil {
			break
		}
		values := merry.Values(gwErr)
		fs := make([]zapcore.Field, 0, len(values))
		for name, value := range values {
			if key, ok := name.(string); ok {
				fs = append(fs, zap.Reflect(key, value))
			}
		}
		logger.Error(merry.Message(gwErr), fs...)
	}

	if err := healthcheck.Shutdown(gw.GracePeriod); err != nil {
		logger.Fatal(merry.Message(err), zap.Error(err))
	}

	if err := debug.Shutdown(gw.GracePeriod); err != nil {
		logger.Fatal(merry.Message(err), zap.Error(err))
	}
}

type logHandler struct {
	logger *zap.Logger
}

func (l logHandler) completion(c context.Context, r *httpx.Request, s httpx.ResponseSnapshot) {
	fs := make([]zapcore.Field, 9, 10)
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
	l.logger.Info("request", fs...)
}

func (l logHandler) error(ctx context.Context, _ *httpx.Request, err merry.Error) {
	l.logger.Error(merry.Message(err), zap.String("request-id", ctx.RequestID()), zap.Error(err))
}

func runAuxiliary(server auxiliary.Server, logger *zap.Logger) {
	if err := server.Listen(); err != nil {
		logger.Error(merry.Message(err), zap.Error(err))
		return
	}
	logger.Info("starting auxiliary server", zap.String("name", server.Name()), zap.String("addr", server.Address()))

	if err := server.Serve(); err != nil {
		logger.Error(merry.Message(err), zap.Error(err))
	}
}
