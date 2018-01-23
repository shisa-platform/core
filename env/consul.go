package env

import (
	"strconv"
	"sync"

	"github.com/ansel1/merry"
	consulapi "github.com/hashicorp/consul/api"
)

//go:generate charlatan -output=./consulselfer_charlatan.go Selfer
type Selfer interface {
	Self() (map[string]map[string]interface{}, error)
}

//go:generate charlatan -output=./consulkvgetter_charlatan.go KVGetter
type KVGetter interface {
	Get(string, *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error)
	List(string, *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error)
}

var _ Selfer = (*consulapi.Agent)(nil)
var _ KVGetter = (*consulapi.KV)(nil)
var _ Provider = (*consulProvider)(nil)

type kvMonitor struct {
	Mon        chan<- Value
	Initial    bool
	LastIndex  uint64
	LastResult Value
	Deleted    bool
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
		return "unknown"
	}
}

type consulProvider struct {
	agent Selfer
	kv    KVGetter

	mux   sync.Mutex
	once  sync.Once
	kvMap map[string]*kvMonitor
}

func NewConsul(c *consulapi.Client) Provider {
	return &consulProvider{c.Agent(), c.KV(), sync.Mutex{}, sync.Once{}, nil}
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

func (p *consulProvider) Monitor(key string, v chan<- Value) {
	p.once.Do(func() {
		p.kvMap = make(map[string]*kvMonitor)
		go p.initMonitor()
	})

	p.mux.Lock()
	p.kvMap[key] = &kvMonitor{Mon: v, Initial: true}
	p.mux.Unlock()

	return
}

func (p *consulProvider) Healthcheck() merry.Error {
	s, err := p.status()
	if err != nil {
		return err
	}

	if s == StatusAlive {
		return nil
	}

	return merry.New("consul agent not alive").WithValue("status", s.String())
}

func (p *consulProvider) status() (status MemberStatus, merr merry.Error) {
	s, err := p.agent.Self()
	if err != nil {
		merr = merry.Wrap(err)
		return
	}

	status, ok := s["Member"]["Status"].(MemberStatus)
	if !ok {
		merr = merry.New("invalid member status for call to agent.Self")
	}

	return
}

func (p *consulProvider) initMonitor() {
	lastIndex := uint64(0)

	for {
		opts := &consulapi.QueryOptions{WaitIndex: lastIndex}
		kvps, meta, err := p.kv.List("", opts)
		if err != nil {
			break
		}

		// Have to keep track of deleted keys
		found := make(map[string]bool)

		lastIndex = meta.LastIndex

		p.mux.Lock()
		for _, kvp := range kvps {
			key := string(kvp.Key)
			kvMon, ok := p.kvMap[key]
			if !ok {
				continue
			}

			found[key] = true

			// If index didn't change, continue
			if kvMon.LastIndex == meta.LastIndex {
				continue
			}

			// Continue if index changed, but the K/V pair didn't
			val := Value{Name: key, Value: string(kvp.Value)}
			if !kvMon.Initial && kvMon.LastResult == val {
				continue
			}

			// Reset if index stops being monotonic
			if meta.LastIndex < kvMon.LastIndex {
				kvMon.LastIndex = 0
			}

			// Handle the updated result
			if !kvMon.Initial {
				kvMon.Mon <- val
			}

			kvMon.Initial = false
			kvMon.LastIndex = meta.LastIndex
			kvMon.LastResult = val
			kvMon.Deleted = false
		}

		// Send blank Value if key is deleted
		for k, v := range p.kvMap {
			if _, ok := found[k]; !ok && !v.Deleted {
				v.Mon <- Value{}
				v.Deleted = true
			}
		}
		p.mux.Unlock()
	}
	return
}
