package sd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"
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
	agent  consulRegistry
	health consulResolver
}

var _ Registrar = &consulSD{}
var _ Resolver = &consulSD{}

func NewConsul(client *consul.Client) *consulSD {
	return &consulSD{
		agent:  client.Agent(),
		health: client.Health(),
	}
}

func (r *consulSD) Register(service string, u *url.URL) merry.Error {
	q := u.Query()

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return merry.Prepend(err, "consul sd: register").Append(u.Host)
	}

	servreg := &consul.AgentServiceRegistration{
		ID:                popQueryString(q, "id"),
		Name:              service,
		Port:              port,
		Address:           u.Hostname(),
		Tags:              popQuerySlice(q, "tag"),
		EnableTagOverride: popQueryBool(q, "enabletagoverride"),
	}

	err = r.agent.ServiceRegister(servreg)
	if err != nil {
		return merry.Prepend(err, "consul sd: register").Append(u.Host)
	}
	return nil
}

func (r *consulSD) Deregister(service string) merry.Error {
	err := r.agent.ServiceDeregister(service)
	if err != nil {
		return merry.Prepend(err, "consul sd: deregister").Append(service)
	}
	return nil
}

func (r *consulSD) Resolve(name string) (nodes []string, merr merry.Error) {
	ses, _, err := r.health.Service(name, "", true, nil)
	if err != nil {
		merr = merry.Prepend(err, "consul sd: resolve").Append(name)
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

func (r *consulSD) AddCheck(service string, u *url.URL) merry.Error {
	q := u.Query()

	acr := &consul.AgentCheckRegistration{
		ID:        popQueryString(q, "id"),
		Name:      fmt.Sprintf("%s-healthcheck", service),
		Notes:     popQueryString(q, "notes"),
		ServiceID: popQueryString(q, "serviceid"),
		AgentServiceCheck: consul.AgentServiceCheck{
			CheckID:       popQueryString(q, "checkid"),
			Timeout:       popQueryString(q, "timeout"),
			Status:        popQueryString(q, "status"),
			TLSSkipVerify: popQueryBool(q, "tlsskipverify"),
		},
	}

	switch u.Scheme {
	case "grpc":
		if len(u.Host) == 0 {
			return merry.New("consul sd: add check: grpc scheme requres host")
		}
		if err := validateKeysExist(q, "interval"); err != nil {
			return err.Prepend("consul sd: add check: grpc scheme")
		}
		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
		acr.AgentServiceCheck.GRPC = u.Host
		acr.AgentServiceCheck.GRPCUseTLS = popQueryBool(q, "grpcusetls")
	case "docker":
		if err := validateKeysExist(q, "dockercontainerid", "args", "interval"); err != nil {
			return err.Prepend("consul sd: add check: docker scheme")
		}
		acr.AgentServiceCheck.DockerContainerID = popQueryString(q, "dockercontainerid")
		acr.AgentServiceCheck.Args = popQuerySlice(q, "args")
		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
		acr.AgentServiceCheck.Shell = popQueryString(q, "shell")
	case "script":
		if err := validateKeysExist(q, "args", "interval"); err != nil {
			return err.Prepend("consul sd: add check: script scheme")
		}
		acr.AgentServiceCheck.Args = popQuerySlice(q, "args")
		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
	case "ttl":
		acr.AgentServiceCheck.TTL = popQueryString(q, "ttl")
	case "tcp":
		if len(u.Host) == 0 {
			return merry.New("consul sd: add check: tcp schema requires host")
		}
		if err := validateKeysExist(q, "interval"); err != nil {
			return err.Prepend("consul sd: add check: tcp scheme")
		}
		acr.AgentServiceCheck.TCP = u.Host
		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
	case "http", "https":
		if len(u.Host) == 0 {
			return merry.New("consul sd: add check: http(s) scheme requries host")
		}
		if err := validateKeysExist(q, "interval"); err != nil {
			return err.Prepend("consul sd: add check: http(s) scheme")
		}

		acr.AgentServiceCheck.Interval = popQueryString(q, "interval")
		acr.AgentServiceCheck.Method = popQueryString(q, "method")
		acr.AgentServiceCheck.Header = q

		u.RawQuery = ""
		urlstr := u.String()

		acr.AgentServiceCheck.HTTP = urlstr
	}

	err := r.agent.CheckRegister(acr)
	if err != nil {
		return merry.Prepend(err, "consul sd: add check")
	}

	return nil
}

func (r *consulSD) RemoveChecks(serviceID string) merry.Error {
	err := r.agent.CheckDeregister(fmt.Sprintf("%s-healthcheck", serviceID))
	if err != nil {
		return merry.Prepend(err, "consul sd: remove checks")
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
	re, ok := v[key]
	if ok {
		delete(v, key)
	}
	return
}

func validateKeysExist(v url.Values, keys ...string) (err merry.Error) {
	err = nil
	for _, key := range keys {
		if _, ok := v[key]; !ok {
			return merry.New("required query parameter missing").Append(key)
		}
	}
	return
}
