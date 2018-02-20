package main

import (
	"net/rpc"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/examples/idp/service"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/sd"
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
	Env      env.Provider
	Resolver sd.Resolver
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
		return nil, merry.Wrap(rpcErr)
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
		return merry.Wrap(rpcErr).WithUserMessage("unable to complete request")
	}
	if !ready {
		return merry.New("not ready").WithUserMessage("not ready")
	}

	return nil
}

func (p *ExampleIdentityProvider) connect() (*rpc.Client, merry.Error) {
	addrs, err := p.resolver.Resolve(p.Name())
	if err != nil {
		return nil, err.WithUserMessage("address service not found")
	}

	if len(addrs) < 1 {
		return nil, merry.New("no healthy hosts")
	}

	client, rpcErr := rpc.DialHTTP("tcp", addrs[0])
	if rpcErr != nil {
		return nil, merry.Wrap(rpcErr).WithUserMessage("unable to connect")
	}

	return client, nil
}
