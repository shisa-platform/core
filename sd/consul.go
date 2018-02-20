package sd

import (
	"fmt"
	"net"
	"net/url"
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

func (r *consulSD) AddCheck(serviceID string, u *url.URL) merry.Error {
	q := u.Query()

	acr := &consul.AgentCheckRegistration{
		ID:        popQueryString(q, "id"),
		Name:      fmt.Sprintf("%s-healthcheck", serviceID),
		Notes:     popQueryString(q, "notes"),
		ServiceID: serviceID,
		AgentServiceCheck: consul.AgentServiceCheck{
			CheckID:           popQueryString(q, "checkid"),
			Args:              popQuerySlice(q, "args"),
			DockerContainerID: popQueryString(q, "dockercontainerid"),
			Shell:             popQueryString(q, "shell"),
			Interval:          popQueryString(q, "interval"),
			Timeout:           popQueryString(q, "timeout"),
			TTL:               popQueryString(q, "ttl"),
			Method:            popQueryString(q, "method"),
			Status:            popQueryString(q, "status"),
			TLSSkipVerify:     popQueryBool(q, "tlsskipverify"),
		},
	}

	enc := q.Encode()

	switch u.Scheme {
	// TODO: wait for https://github.com/hashicorp/consul/commit/c3e94970a09db21b1a3de947ae28577980a18161
	// to get released before handling GRPC
	case "grpc":
		//acr.AgentServiceCheck.GRPC = fmt.Sprintf("grpc://%s%s", u.Host, u.Path)
	case "tcp":
		acr.AgentServiceCheck.TCP = u.Host
	case "http", "https":
		var s string
		if len(enc) > 0 {
			s = fmt.Sprintf("%s://%s%s?%s", u.Scheme, u.Host, u.Path, enc)
		} else {
			s = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
		}
		acr.AgentServiceCheck.HTTP = s
	}

	acr.Header = q

	err := r.client.Agent().CheckRegister(acr)
	if err != nil {
		return merry.Wrap(err)
	}

	return nil
}

func (r *consulSD) ClearChecks(serviceID string) merry.Error {
	err := r.client.Agent().CheckDeregister(fmt.Sprintf("%s-healthcheck", serviceID))
	if err != nil {
		return merry.Wrap(err)
	}
	return nil
}

func popQueryString(v url.Values, key string) (re string) {
	val, ok := v[key]
	if ok {
		delete(v, key)
		re = val[0]
	}
	return
}

func popQueryBool(v url.Values, key string) (re bool) {
	val, ok := v[key]
	if ok {
		delete(v, key)
	}
	if len(val) > 0 {
		re, _ = strconv.ParseBool(val[0])
	}
	return
}

func popQuerySlice(v url.Values, key string) (re []string) {
	val, ok := v[key]
	if ok {
		delete(v, key)
		re = val
	}
	return
}
