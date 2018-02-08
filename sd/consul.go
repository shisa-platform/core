package sd

import (
	"fmt"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"
)

type consulRegistrar struct {
	client  *consul.Client
	servreg *consul.AgentServiceRegistration
}

func NewConsulRegistrar(client *consul.Client, servreg *consul.AgentServiceRegistration) *consulRegistrar {
	return &consulRegistrar{
		client:  client,
		servreg: servreg,
	}
}

func (r *consulRegistrar) Register() merry.Error {
	e := r.client.Agent().ServiceRegister(r.servreg)
	if e != nil {
		return merry.Wrap(e)
	}
	return nil
}

func (r *consulRegistrar) Deregister() merry.Error {
	e := r.client.Agent().ServiceDeregister(r.servreg.ID)
	if e != nil {
		return merry.Wrap(e)
	}
	return nil
}

type consulResolver struct {
	client      *consul.Client
	passingOnly bool
}

func NewConsulResolver(client *consul.Client, passingOnly bool) *consulResolver {
	return &consulResolver{client, passingOnly}
}

func (r *consulResolver) Resolve(name string) (nodes []string, merr merry.Error) {
	ses, _, err := r.client.Health().Service(name, "", r.passingOnly, nil)
	if err != nil {
		merr = merry.Wrap(err)
		return
	}

	nodes = make([]string, len(ses))
	for i, s := range ses {
		addr := s.Node.Address
		if s.Service.Address != "" {
			addr = s.Service.Address
		}
		nodes[i] = fmt.Sprintf("%s:%d", addr, s.Service.Port)
	}
	return
}
