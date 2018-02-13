package sd

import (
	"net"
	"strconv"

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
	client *consul.Client
}

func NewConsulResolver(client *consul.Client) *consulResolver {
	return &consulResolver{client}
}

func (r *consulResolver) Resolve(name string, passingOnly bool) (nodes []string, merr merry.Error) {
	ses, _, err := r.client.Health().Service(name, "", passingOnly, nil)
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
		nodes[i] = net.JoinHostPort(addr, strconv.Itoa(s.Service.Port))
	}
	return
}
