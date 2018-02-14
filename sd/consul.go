package sd

import (
	"net"
	"strconv"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"
)

type consulSD struct {
	client *consul.Client
}

var _ Registrar = &consulSD{}
var _ Resolver = &consulSD{}

func NewConsul(client *consul.Client) *consulSD {
	return &consulSD{
		client: client,
	}
}

func (r *consulSD) Register(name, addr string) merry.Error {
	address, sport, err := net.SplitHostPort(addr)
	if err != nil {
		return merry.Errorf("splitting addr/port: %v", err)
	}

	port, err := strconv.Atoi(sport)
	if err != nil {
		return merry.Errorf("parsing port: %v", err)
	}

	servreg := &consul.AgentServiceRegistration{
		ID:      name,
		Name:    name,
		Port:    port,
		Address: address,
	}

	e := r.client.Agent().ServiceRegister(servreg)
	if e != nil {
		return merry.Wrap(e)
	}
	return nil
}

func (r *consulSD) Deregister(name string) merry.Error {
	e := r.client.Agent().ServiceDeregister(name)
	if e != nil {
		return merry.Wrap(e)
	}
	return nil
}

func (r *consulSD) Resolve(name string) (nodes []string, merr merry.Error) {
	ses, _, err := r.client.Health().Service(name, "", true, nil)
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
