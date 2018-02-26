package sd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"

	"github.com/percolate/shisa/lb"
)

//go:generate charlatan -output=./consulregistry_charlatan.go consulRegistry
type consulRegistry interface {
	ServiceRegister(*consul.AgentServiceRegistration) error
	ServiceDeregister(string) error
	CheckRegister(*consul.AgentCheckRegistration) error
	CheckDeregister(string) error
}

//go:generate charlatan -output=./consulresolver_charlatan.go consulResolver
type consulResolver interface {
	Service(service, tag string, passingOnly bool, q *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error)
}

type consulSD struct {
	agent    consulRegistry
	health   consulResolver
	balancer lb.Balancer
}

var _ Registrar = &consulSD{}
var _ Resolver = &consulSD{}

func NewConsulLB(client *consul.Client, b lb.Balancer) *consulSD {
	return &consulSD{
		agent:    client.Agent(),
		health:   client.Health(),
		balancer: b,
	}
}

func NewConsul(client *consul.Client) *consulSD {
	return &consulSD{
		agent:  client.Agent(),
		health: client.Health(),
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

	e := r.agent.ServiceRegister(servreg)
	if e != nil {
		return merry.Wrap(e)
	}
	return nil
}

func (r *consulSD) Deregister(serviceID string) merry.Error {
	e := r.agent.ServiceDeregister(serviceID)
	if e != nil {
		return merry.Wrap(e)
	}
	return nil
}

func (r *consulSD) Resolve(name string) (nodes []string, merr merry.Error) {
	ses, _, err := r.health.Service(name, "", true, nil)
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
			CheckID:       popQueryString(q, "checkid"),
			Timeout:       popQueryString(q, "timeout"),
			Status:        popQueryString(q, "status"),
			TLSSkipVerify: popQueryBool(q, "tlsskipverify"),
		},
	}

	switch u.Scheme {
	// TODO: wait for https://github.com/hashicorp/consul/commit/c3e94970a09db21b1a3de947ae28577980a18161
	// to get released before handling GRPC
	case "grpc":
		if len(u.Host) == 0 {
			return merry.New("url Host field required for gRPC check")
		}
		if err := validateKeysExist(q, "interval"); err != nil {
			return err
		}
		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
		// acr.AgentServiceCheck.GRPC = u.Host
		// acr.AgentServiceCheck.GRPCUseTLS = popQueryBool(q, "grpcusetls")
	case "docker":
		if err := validateKeysExist(q, "dockercontainerid", "args", "interval"); err != nil {
			return err
		}
		acr.AgentServiceCheck.DockerContainerID = popQueryString(q, "dockercontainerid")
		acr.AgentServiceCheck.Args = popQuerySlice(q, "args")
		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
		acr.AgentServiceCheck.Shell = popQueryString(q, "shell")
	case "script":
		if err := validateKeysExist(q, "args", "interval"); err != nil {
			return err
		}
		acr.AgentServiceCheck.Args = popQuerySlice(q, "args")
		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
	case "ttl":
		acr.AgentServiceCheck.TTL = popQueryString(q, "ttl")
	case "tcp":
		if len(u.Host) == 0 {
			return merry.New("url Host field required for TCP check")
		}
		if err := validateKeysExist(q, "interval"); err != nil {
			return err
		}
		acr.AgentServiceCheck.TCP = u.Host
		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
	case "http", "https":
		urlstr := u.String()
		if len(urlstr) == 0 {
			return merry.New("non-empty url.String required for HTTP(S) check")
		}
		if err := validateKeysExist(q, "interval"); err != nil {
			return err
		}
		u.RawQuery = ""
		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
		acr.AgentServiceCheck.Method = popQueryString(q, "method")
		acr.AgentServiceCheck.Header = q
		acr.AgentServiceCheck.HTTP = urlstr
	}

	err := r.agent.CheckRegister(acr)
	if err != nil {
		return merry.Wrap(err)
	}

	return nil
}

func (r *consulSD) RemoveChecks(serviceID string) merry.Error {
	err := r.agent.CheckDeregister(fmt.Sprintf("%s-healthcheck", serviceID))
	if err != nil {
		return merry.Wrap(err)
	}
	return nil
}

func (r *consulSD) Balance(name string) (string, merry.Error) {
	addrs, err := r.Resolve(name)
	if err != nil {
		return "", err
	}
	addr, err := r.balancer.Balance(addrs)
	if err != nil {
		return "", err
	}
	return addr, nil
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

func validateKeysExist(v url.Values, keys ...string) (err merry.Error) {
	err = nil
	for _, k := range keys {
		if _, ok := v[k]; !ok {
			return merry.New("key/value pair not found in params").WithValue("key", k)
		}
	}
	return
}
