package main

import (
	"net/rpc"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/examples/idp/service"
	"github.com/percolate/shisa/lb"
	"github.com/percolate/shisa/models"
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
	client, err := p.connect()
	if err != nil {
		return nil, err
	}

	message := idp.Message{RequestID: ctx.RequestID(), Value: credentials}
	var userID string
	rpcErr := client.Call("Idp.AuthenticateToken", &message, &userID)
	if rpcErr != nil {
		return nil, merry.Prepend(rpcErr, "authenticate")
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
