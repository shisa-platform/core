package sd

import (
	"fmt"
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

func (r *consulSD) Register(serviceID, addr string) merry.Error {
	address, sport, err := net.SplitHostPort(addr)
	if err != nil {
		return merry.Errorf("splitting addr/port: %v", err)
	}

	port, err := strconv.Atoi(sport)
	if err != nil {
		return merry.Errorf("parsing port: %v", err)
	}

	servreg := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceID,
		Port:    port,
		Address: address,
	}

	e := r.client.Agent().ServiceRegister(servreg)
	if e != nil {
		return merry.Wrap(e)
	}
	return nil
}

func (r *consulSD) Deregister(serviceID string) merry.Error {
	e := r.client.Agent().ServiceDeregister(serviceID)
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

func (r *consulSD) AddHealthcheck(serviceID, url string) merry.Error {
	acr := &consul.AgentCheckRegistration{
		ID:   fmt.Sprintf("%s-healthcheck", serviceID),
		Name: fmt.Sprintf("%s-healthcheck", serviceID),
		AgentServiceCheck: consul.AgentServiceCheck{
			TCP:      url,
			Interval: "5s",
		},
	}

	err := r.client.Agent().CheckRegister(acr)
	if err != nil {
		return merry.Wrap(err)
	}

	return nil
}

func (r *consulSD) RemoveHealthcheck(serviceID, url string) merry.Error {
	err := r.client.Agent().CheckDeregister(fmt.Sprintf("%s-healthcheck", serviceID))
	if err != nil {
		return merry.Wrap(err)
	}
	return nil
}
