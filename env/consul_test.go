package env

import (
	stdctx "context"
	"sync"
	"testing"

	"github.com/ansel1/merry"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
)

var (
	missingConsul = newFakeConsul([]byte(""), merry.New("missing"))
	emptyConsul   = newFakeConsul([]byte(""), nil)
	defaultConsul = newFakeConsul(defaultVal, nil)
	intConsul     = newFakeConsul(defaultIntVal, nil)
	boolConsul    = newFakeConsul(defaultBoolVal, nil)
)

func newFakeConsul(val []byte, err error) *ConsulProvider {
	return &ConsulProvider{
		agent: &Fakeselfer{},
		kv: &FakekvGetter{
			GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
				if err != nil {
					return nil, nil, err
				}
				return &consulapi.KVPair{Value: val}, nil, nil
			},
		},
	}
}

func TestMemberStatusString(t *testing.T) {
	tests := []struct {
		status memberStatus
		str    string
	}{
		{statusNone, "none"},
		{statusAlive, "alive"},
		{statusLeaving, "leaving"},
		{statusLeft, "left"},
		{statusFailed, "failed"},
		{999999, "unknown"},
	}
	for _, test := range tests {
		t.Run(test.str, func(t *testing.T) {
			assert.Equal(t, test.str, test.status.String())
		})
	}
}

func TestNewConsul(t *testing.T) {
	ac := &consulapi.Client{}
	prefix := "environment/prefix/"

	c := NewConsul(ac, prefix)

	assert.NotNil(t, c)
	assert.NotNil(t, c.agent)
	assert.NotNil(t, c.kv)
	assert.Equal(t, prefix, c.prefix)
}

func TestConsulProviderGet(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultVal}, nil, nil
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.Get(defaultKey)

	assert.Equal(t, string(defaultVal), r)
	assert.NoError(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetError(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return nil, nil, merry.New("get error")
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetEmpty(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: []byte("")}, nil, nil
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetMissing(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return nil, nil, nil
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	assert.True(t, merry.Is(err, NameNotSet))
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetInt(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultIntVal}, nil, nil
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetInt(defaultKey)

	assert.Equal(t, defaultInt, r)
	assert.NoError(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetIntError(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return nil, nil, merry.New("get error")
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetInt(defaultKey)

	assert.Zero(t, r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetIntParseFailure(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultVal}, nil, nil
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetInt(defaultKey)

	assert.Zero(t, r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetIntMissing(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return nil, nil, nil
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetInt(defaultKey)

	assert.Zero(t, r)
	assert.Error(t, err)
	assert.True(t, merry.Is(err, NameNotSet))
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetBool(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultBoolVal}, nil, nil
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetBool(defaultKey)

	assert.Equal(t, true, r)
	assert.NoError(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetBoolError(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return nil, nil, merry.New("get error")
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetBool(defaultKey)

	assert.False(t, r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetBoolParseFailure(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultVal}, nil, nil
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetBool(defaultKey)

	assert.False(t, r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetBoolMissing(t *testing.T) {
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return nil, nil, nil
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetBool(defaultKey)

	assert.False(t, r)
	assert.Error(t, err)
	assert.True(t, merry.Is(err, NameNotSet))
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderMonitorNoChange(t *testing.T) {
	i := uint64(10)
	z := 0
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { z++ }()
			if z < 3 {
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i}, nil
			} else {
				return nil, nil, merry.New("stop")
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}
	v := make(chan Value, 5)

	c.Monitor(defaultKey, v)

	assert.Empty(t, v)
}

func TestConsulProviderMonitorIndexChange(t *testing.T) {
	i := uint64(10)
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { i++ }()
			if i < 13 {
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			} else {
				return nil, nil, merry.New("stop")
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
		mux:   sync.Mutex{},
		once:  sync.Once{},
	}
	v := make(chan Value, 5)

	c.Monitor(defaultKey, v)

	assert.Empty(t, v)
}

func TestConsulProviderMonitorChange(t *testing.T) {
	i := uint64(10)
	newVal := "NEW_VAL"
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { i++ }()
			switch i {
			case 10:
				// First value is not written to channel - Monitor only sends changed vals to channel
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 11:
				return []*consulapi.KVPair{{Key: defaultKey, Value: []byte(newVal)}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 12:
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			default:
				return nil, nil, merry.New("stop")
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
		mux:   sync.Mutex{},
		once:  sync.Once{},
	}
	v := make(chan Value, 2)

	c.Monitor(defaultKey, v)

	v1 := <-v
	v2 := <-v

	expectedV1 := Value{defaultKey, newVal}
	expectedV2 := Value{defaultKey, string(defaultVal)}

	assert.Equal(t, expectedV1, v1)
	assert.Equal(t, expectedV2, v2)
}

func TestConsulProviderMonitorIndexReset(t *testing.T) {
	i := uint64(10)
	newVal := "NEW_VAL"
	s := &Fakeselfer{}

	swap := true
	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { i++ }()

			if i == 12 && swap {
				i = 10
				swap = false
			}

			switch i {
			case 10:
				// First value is not written to channel - Monitor only sends changed vals to channel
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 11:
				return []*consulapi.KVPair{{Key: defaultKey, Value: []byte(newVal)}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 12:
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			default:
				return nil, nil, merry.New("stop")
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}
	v := make(chan Value, 5)

	c.Monitor(defaultKey, v)

	v1 := <-v
	v2 := <-v
	v3 := <-v
	v4 := <-v

	expectedNew := Value{defaultKey, newVal}
	expectedDefault := Value{defaultKey, string(defaultVal)}

	assert.Equal(t, expectedNew, v1)
	assert.Equal(t, expectedDefault, v2)
	assert.Equal(t, expectedNew, v3)
	assert.Equal(t, expectedDefault, v4)
}

func TestConsulProviderMonitorMultipleKeys(t *testing.T) {
	i := uint64(10)
	anyKey := "ANY"
	newVal := "NEW_VAL"
	s := &Fakeselfer{}
	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { i++ }()
			switch i {
			case 10:
				// First value is not written to channel - Monitor only sends changed vals to channel
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: anyKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 11:
				return []*consulapi.KVPair{{Key: defaultKey, Value: []byte(newVal)}, {Key: anyKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 12:
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: anyKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			default:
				return nil, nil, merry.New("stop")
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}
	v := make(chan Value, 2)

	c.Monitor(defaultKey, v)

	v1 := <-v
	v2 := <-v

	expectedNew := Value{defaultKey, newVal}
	expectedDefault := Value{defaultKey, string(defaultVal)}

	assert.Equal(t, expectedNew, v1)
	assert.Equal(t, expectedDefault, v2)
}

func TestConsulProviderMonitorMultipleMonitor(t *testing.T) {
	i := uint64(10)
	anyKey := "ANY"
	newVal := "NEW_VAL"
	s := &Fakeselfer{}

	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { i++ }()

			switch i {
			case 10:
				// First value is not written to channel - Monitor only sends changed vals to channel
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: anyKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 11:
				return []*consulapi.KVPair{{Key: defaultKey, Value: []byte(newVal)}, {Key: anyKey, Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 12:
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: anyKey, Value: []byte(newVal)}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			default:
				return nil, nil, merry.New("stop")
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
		mux:   sync.Mutex{},
		once:  sync.Once{},
	}

	v := make(chan Value, 2)
	z := make(chan Value, 1)

	c.Monitor(defaultKey, v)
	c.Monitor(anyKey, z)

	v1 := <-v
	v2 := <-v

	expectedNew := Value{defaultKey, newVal}
	expectedDefault := Value{defaultKey, string(defaultVal)}
	expectedAny := Value{anyKey, newVal}

	assert.Equal(t, expectedNew, v1)
	assert.Equal(t, expectedDefault, v2)
	assert.Equal(t, expectedAny, <-z)
}

func TestConsulProviderMonitorDeletedKey(t *testing.T) {
	i := uint64(10)
	s := &Fakeselfer{}

	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { i++ }()

			switch i {
			case 10:
				// First value is not written to channel - Monitor only sends changed vals to channel
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: "ANY", Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 11:
				return []*consulapi.KVPair{}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 12:
				return []*consulapi.KVPair{}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			default:
				return nil, nil, merry.New("stop")
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	v := make(chan Value, 1)

	c.Monitor(defaultKey, v)

	expected := Value{}

	assert.Equal(t, expected, <-v)
}

func TestConsulProviderMonitorRevenant(t *testing.T) {
	i := uint64(10)
	s := &Fakeselfer{}

	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { i++ }()

			switch i {
			case 10:
				// First value is not written to channel - Monitor only sends changed vals to channel
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: "ANY", Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 11:
				return []*consulapi.KVPair{}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 12:
				return []*consulapi.KVPair{}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			case 13:
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: "ANY", Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			default:
				return nil, nil, merry.New("stop")
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
		mux:   sync.Mutex{},
		once:  sync.Once{},
	}

	v := make(chan Value, 2)

	c.Monitor(defaultKey, v)

	e1 := Value{}
	e2 := Value{defaultKey, string(defaultVal)}

	assert.Equal(t, e1, <-v)
	assert.Equal(t, e2, <-v)
}

func TestConsulProviderMonitorErrorHandler(t *testing.T) {
	e := merry.New("stop")
	var called error

	calledCh := make(chan struct{})

	errorHandler := func(err error) {
		called = err
		close(calledCh)
		return
	}

	s := &Fakeselfer{}

	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			return nil, nil, e
		},
	}
	c := &ConsulProvider{
		ErrorHandler: errorHandler,

		agent: s,
		kv:    kvg,
	}

	v := make(chan Value)

	c.Monitor(defaultKey, v)
	<-calledCh

	assert.Equal(t, e, called)
}

func TestConsulProviderShutdown(t *testing.T) {
	i := uint64(10)
	s := &Fakeselfer{}

	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { i++ }()

			switch i {
			case i % 2:
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: "ANY", Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			default:
				return []*consulapi.KVPair{}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	v := make(chan Value, 2)

	c.Monitor(defaultKey, v)
	<-v

	c.Shutdown()
	_, ok := <-c.stopCh

	c.Shutdown() // should be a noop

	assert.True(t, c.stop)
	assert.False(t, ok)
}

func TestConsulProviderShutdownNotMonitoring(t *testing.T) {
	s := &Fakeselfer{}

	kvg := &FakekvGetter{}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}
	c.Shutdown() // should be a noop

	assert.Nil(t, c.stopCh)
	assert.False(t, c.IsMonitoring())
}

func TestConsulProviderIsMonitoring(t *testing.T) {
	i := uint64(10)
	s := &Fakeselfer{}

	kvg := &FakekvGetter{
		ListHook: func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
			defer func() { i++ }()

			switch i {
			case i % 2:
				return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: "ANY", Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			default:
				return []*consulapi.KVPair{}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
			}
		},
	}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	v := make(chan Value)

	c.Monitor(defaultKey, v)
	m1 := c.IsMonitoring()
	<-v

	c.Shutdown()
	_, ok := <-c.stopCh
	m2 := c.IsMonitoring()

	c.Shutdown() // should be a noop

	assert.True(t, c.stop)
	assert.False(t, ok)
	assert.True(t, m1)
	assert.False(t, m2)
}

func TestConsulProviderShutdownAfterListCalled(t *testing.T) {
	i := uint64(10)
	s := &Fakeselfer{}
	shut := make(chan struct{})

	kvg := &FakekvGetter{}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	kvg.ListHook = func(s string, options *consulapi.QueryOptions) (consulapi.KVPairs, *consulapi.QueryMeta, error) {
		defer func() {
			c.Shutdown()
			close(shut)
		}()
		defer func() { i++ }()

		return []*consulapi.KVPair{{Key: defaultKey, Value: defaultVal}, {Key: "ANY", Value: defaultVal}}, &consulapi.QueryMeta{LastIndex: i + 1}, nil
	}

	v := make(chan Value)
	c.Monitor(defaultKey, v)
	_, ok := <-shut

	assert.False(t, ok)
	assert.False(t, c.IsMonitoring())
}

func TestConsulProviderHealthcheck(t *testing.T) {
	s := &Fakeselfer{
		SelfHook: func() (map[string]map[string]interface{}, error) {
			m := make(map[string]map[string]interface{})
			m["Member"] = make(map[string]interface{})
			m["Member"]["Status"] = statusAlive
			return m, nil
		},
	}
	kvg := &FakekvGetter{}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	ctx := context.New(stdctx.Background())
	err := c.Healthcheck(ctx)

	assert.NoError(t, err)
	s.AssertSelfCalledOnce(t)
}

func TestConsulProviderHealthcheckStatusError(t *testing.T) {
	s := &Fakeselfer{
		SelfHook: func() (map[string]map[string]interface{}, error) {
			return nil, merry.New("self failure")
		},
	}
	kvg := &FakekvGetter{}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	ctx := context.New(stdctx.Background())
	err := c.Healthcheck(ctx)

	assert.Error(t, err)
	s.AssertSelfCalledOnce(t)
}

func TestConsulProviderHealthcheckNotAlive(t *testing.T) {
	s := &Fakeselfer{
		SelfHook: func() (map[string]map[string]interface{}, error) {
			m := make(map[string]map[string]interface{})
			m["Member"] = make(map[string]interface{})
			m["Member"]["Status"] = statusFailed
			return m, nil
		},
	}
	kvg := &FakekvGetter{}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	ctx := context.New(stdctx.Background())
	err := c.Healthcheck(ctx)

	assert.Error(t, err)
	s.AssertSelfCalledOnce(t)
}

func TestConsulProviderStatusUnparseable(t *testing.T) {
	s := &Fakeselfer{
		SelfHook: func() (map[string]map[string]interface{}, error) {
			m := make(map[string]map[string]interface{})
			m["Member"] = make(map[string]interface{})
			m["Member"]["Status"] = "unparseable"
			return m, nil
		},
	}
	kvg := &FakekvGetter{}
	c := &ConsulProvider{
		agent: s,
		kv:    kvg,
	}

	ctx := context.New(stdctx.Background())
	err := c.Healthcheck(ctx)

	assert.Error(t, err)
	s.AssertSelfCalledOnce(t)
}
