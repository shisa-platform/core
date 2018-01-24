package env

import (
	"context"
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
	ch         chan<- Value
	init       bool
	lastIndex  uint64
	lastResult Value
	del        bool
}

type MemberStatus int

type ErrorHandler func(error)

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

	stop       bool
	stopCh     chan struct{}
	stopLock   sync.Mutex
	cancelFunc context.CancelFunc

	ErrorHandler ErrorHandler
}

func NewConsul(c *consulapi.Client) *consulProvider {
	return &consulProvider{
		agent: c.Agent(),
		kv:    c.KV(),
	}
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
		p.stopCh = make(chan struct{})
		go p.monitorLoop()
	})

	p.mux.Lock()
	p.kvMap[key] = &kvMonitor{ch: v, init: true}
	p.mux.Unlock()

	return
}

func (p *consulProvider) Shutdown() {
	if p.stopCh == nil {
		return
	}

	p.stopLock.Lock()
	defer p.stopLock.Unlock()

	if p.stop {
		return
	}
	p.stop = true

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	close(p.stopCh)

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

func (p *consulProvider) monitorLoop() {
	lastIndex := uint64(0)

	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFunc = cancel

	for !p.shouldStop() {
		opts := &consulapi.QueryOptions{WaitIndex: lastIndex}

		kvps, meta, err := p.kv.List("", opts.WithContext(ctx))
		if err != nil {
			if p.ErrorHandler != nil {
				p.ErrorHandler(err)
			}
			break
		}

		// Check if we should terminate since the function
		// could have blocked for a while
		if p.shouldStop() {
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
			if kvMon.lastIndex == meta.LastIndex {
				continue
			}

			// Continue if index changed, but the K/V pair didn't
			val := Value{Name: key, Value: string(kvp.Value)}
			if !kvMon.init && kvMon.lastResult == val {
				continue
			}

			// Reset if index stops being monotonic
			if meta.LastIndex < kvMon.lastIndex {
				kvMon.lastIndex = 0
			}

			// Handle the updated result
			if !kvMon.init {
				kvMon.ch <- val
			}

			kvMon.init = false
			kvMon.lastIndex = meta.LastIndex
			kvMon.lastResult = val
			kvMon.del = false
		}

		// Send blank Value if key is deleted
		for k, v := range p.kvMap {
			if _, ok := found[k]; !ok && !v.del {
				val := Value{}
				v.ch <- val
				v.lastResult = val
				v.del = true
			}
		}
		p.mux.Unlock()
	}
	return
}

func (p *consulProvider) shouldStop() bool {
	select {
	case <-p.stopCh:
		return true
	default:
		return false
	}
}

func (p *consulProvider) IsMonitoring() bool {
	if p.stopCh == nil {
		return false
	}

	p.stopLock.Lock()
	defer p.stopLock.Unlock()
	return !p.stop
}
