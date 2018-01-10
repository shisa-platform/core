package env

import (
	"fmt"
	"strconv"

	"github.com/ansel1/merry"
	consulapi "github.com/hashicorp/consul/api"
)

type ConsulClient interface {
	Agent() *consulapi.Agent
	KV() *consulapi.KV
}

type MemberStatus int

const (
	StatusNone MemberStatus = iota
	StatusAlive
	StatusLeaving
	StatusLeft
	StatusFailed
)

func (s MemberStatus) String() string {
	switch s {
	case StatusNone:
		return "none"
	case StatusAlive:
		return "alive"
	case StatusLeaving:
		return "leaving"
	case StatusLeft:
		return "left"
	case StatusFailed:
		return "failed"
	default:
		return fmt.Sprintf("unknown MemberStatus: %d", s)
	}
}

type consulProvider struct {
	agent *consulapi.Agent
	kv    *consulapi.KV
}

func NewConsul(c ConsulClient) Provider {
	return &consulProvider{c.Agent(), c.KV()}
}

func (p *consulProvider) Get(name string) (string, merry.Error) {
	kvp, _, err := p.kv.Get(name, nil)
	if err != nil {
		return "", merry.Wrap(err).WithValue("name", name)
	}

	r := string(kvp.Value)
	if r == "" {
		return "", NameEmpty.Here().WithValue("name", name)
	}

	return r, nil
}

func (p *consulProvider) GetInt(name string) (int, merry.Error) {
	kvp, _, err := p.kv.Get(name, nil)
	if err != nil {
		return 0, merry.Wrap(err).WithValue("name", name)
	}

	r, err := strconv.Atoi(string(kvp.Value))
	if err != nil {
		return 0, merry.Wrap(err).WithValue("name", name)
	}

	return r, nil
}

func (p *consulProvider) GetBool(name string) (bool, merry.Error) {
	kvp, _, err := p.kv.Get(name, nil)
	if err != nil {
		return false, merry.Wrap(err).WithValue("name", name)
	}

	r, err := strconv.ParseBool(string(kvp.Value))
	if err != nil {
		return false, merry.Wrap(err).WithValue("name", name)
	}

	return r, nil
}

func (p *consulProvider) Monitor(string, <-chan Value) {
	// N.B. do nothing, not implemented
}

func (p *consulProvider) Healthcheck() merry.Error {
	s, err := p.status()
	if err != nil {
		return err
	}

	if s == StatusAlive {
		return nil
	}

	return merry.Errorf("bad consul agent status: %s", s)
}

func (p *consulProvider) status() (status MemberStatus, merr merry.Error) {
	s, err := p.agent.Self()
	if err != nil {
		merr = merry.Wrap(err)
	}

	status, ok := s["Member"]["Status"].(MemberStatus)
	if !ok {
		merr = merry.New("invalid member status for call to agent.Self")
	}

	return
}
