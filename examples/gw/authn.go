package main

import (
	"net/rpc"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/examples/idp/service"
	"github.com/shisa-platform/core/lb"
	"github.com/shisa-platform/core/models"
)

type simpleUser struct {
	ident string
}

func (u simpleUser) ID() string {
	return u.ident
}

func (u simpleUser) String() string {
	return u.ident
}

type ExampleIdentityProvider struct {
	Balancer lb.Balancer
}

func (p *ExampleIdentityProvider) Authenticate(ctx context.Context, credentials string) (models.User, merry.Error) {
	var span opentracing.Span
	if ctx.Span() != nil {
		span = ctx.StartSpan("IdentityProvider.Authenticate")
		defer span.Finish()
	} else {
		tracer := opentracing.NoopTracer{}
		span = tracer.StartSpan("noop")
	}

	client, err := p.connect()
	if err != nil {
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
		return nil, err
	}

	message := idp.Message{
		RequestID: ctx.RequestID(),
		Value:     credentials,
		Metadata:  make(map[string]string),
	}

	carrier := opentracing.TextMapCarrier(message.Metadata)
	if err := span.Tracer().Inject(span.Context(), opentracing.TextMap, carrier); err != nil {
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
		return nil, merry.Prepend(err, "authenticate")
	}

	var userID string
	if err := client.Call("Idp.AuthenticateToken", &message, &userID); err != nil {
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
		return nil, merry.Prepend(err, "authenticate")
	}
	if userID == "" {
		return nil, nil
	}

	return simpleUser{ident: userID}, nil
}

func (p *ExampleIdentityProvider) Name() string {
	return "idp"
}

func (p *ExampleIdentityProvider) Healthcheck(ctx context.Context) merry.Error {
	client, err := p.connect()
	if err != nil {
		return err
	}

	var ready bool
	arg := ctx.RequestID()
	rpcErr := client.Call("Idp.Healthcheck", &arg, &ready)
	if rpcErr != nil {
		return merry.Prepend(rpcErr, "healthcheck")
	}
	if !ready {
		return merry.New("not ready")
	}

	return nil
}

func (p *ExampleIdentityProvider) connect() (*rpc.Client, merry.Error) {
	addr, err := p.Balancer.Balance(p.Name())
	if err != nil {
		return nil, err.Prepend("connect")
	}

	client, rpcErr := rpc.DialHTTP("tcp", addr)
	if rpcErr != nil {
		return nil, merry.Prepend(rpcErr, "connect")
	}

	return client, nil
}
