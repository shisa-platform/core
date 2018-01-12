package env

import (
	"testing"

	"github.com/ansel1/merry"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

const defaultKey = "DEFAULT_KEY"

var (
	defaultVal     = []byte("DEFAULT_VAL")
	defaultIntVal  = []byte("1")
	defaultInt     = 1
	defaultBoolVal = []byte("true")
)

func TestMemberStatusString(t *testing.T) {
	tests := []struct {
		status MemberStatus
		str    string
	}{
		{StatusNone, "none"},
		{StatusAlive, "alive"},
		{StatusLeaving, "leaving"},
		{StatusLeft, "left"},
		{StatusFailed, "failed"},
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

	c := NewConsul(ac)
	cs := c.(*consulProvider)

	assert.NotNil(t, c)
	assert.NotNil(t, cs.agent)
	assert.NotNil(t, cs.kv)
}

func TestConsulProviderGet(t *testing.T) {
	s := &FakeSelfer{}
	kvg := &FakeKVGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultVal}, nil, nil
		},
	}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.Get(defaultKey)

	assert.Equal(t, string(defaultVal), r)
	assert.NoError(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetError(t *testing.T) {
	s := &FakeSelfer{}
	kvg := &FakeKVGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return nil, nil, merry.New("get error")
		},
	}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetEmpty(t *testing.T) {
	s := &FakeSelfer{}
	kvg := &FakeKVGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: []byte("")}, nil, nil
		},
	}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetInt(t *testing.T) {
	s := &FakeSelfer{}
	kvg := &FakeKVGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultIntVal}, nil, nil
		},
	}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetInt(defaultKey)

	assert.Equal(t, defaultInt, r)
	assert.NoError(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetIntError(t *testing.T) {
	s := &FakeSelfer{}
	kvg := &FakeKVGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return nil, nil, merry.New("get error")
		},
	}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetInt(defaultKey)

	assert.Zero(t, r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetIntParseFailure(t *testing.T) {
	s := &FakeSelfer{}
	kvg := &FakeKVGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultVal}, nil, nil
		},
	}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetInt(defaultKey)

	assert.Zero(t, r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetBool(t *testing.T) {
	s := &FakeSelfer{}
	kvg := &FakeKVGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultBoolVal}, nil, nil
		},
	}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetBool(defaultKey)

	assert.Equal(t, true, r)
	assert.NoError(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetBoolError(t *testing.T) {
	s := &FakeSelfer{}
	kvg := &FakeKVGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return nil, nil, merry.New("get error")
		},
	}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetBool(defaultKey)

	assert.False(t, r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderGetBoolParseFailure(t *testing.T) {
	s := &FakeSelfer{}
	kvg := &FakeKVGetter{
		GetHook: func(s string, options *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
			return &consulapi.KVPair{Value: defaultVal}, nil, nil
		},
	}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	r, err := c.GetBool(defaultKey)

	assert.False(t, r)
	assert.Error(t, err)
	kvg.AssertGetCalledOnceWith(t, defaultKey, nil)
}

func TestConsulProviderHealthcheck(t *testing.T) {
	s := &FakeSelfer{
		SelfHook: func() (map[string]map[string]interface{}, error) {
			m := make(map[string]map[string]interface{})
			m["Member"] = make(map[string]interface{})
			m["Member"]["Status"] = StatusAlive
			return m, nil
		},
	}
	kvg := &FakeKVGetter{}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	err := c.Healthcheck()

	assert.NoError(t, err)
	s.AssertSelfCalledOnce(t)
}

func TestConsulProviderHealthcheckStatusError(t *testing.T) {
	s := &FakeSelfer{
		SelfHook: func() (map[string]map[string]interface{}, error) {
			return nil, merry.New("self failure")
		},
	}
	kvg := &FakeKVGetter{}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	err := c.Healthcheck()

	assert.Error(t, err)
	s.AssertSelfCalledOnce(t)
}

func TestConsulProviderHealthcheckNotAlive(t *testing.T) {
	s := &FakeSelfer{
		SelfHook: func() (map[string]map[string]interface{}, error) {
			m := make(map[string]map[string]interface{})
			m["Member"] = make(map[string]interface{})
			m["Member"]["Status"] = StatusFailed
			return m, nil
		},
	}
	kvg := &FakeKVGetter{}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	err := c.Healthcheck()

	assert.Error(t, err)
	s.AssertSelfCalledOnce(t)
}

func TestConsulProviderStatusUnparseable(t *testing.T) {
	s := &FakeSelfer{
		SelfHook: func() (map[string]map[string]interface{}, error) {
			m := make(map[string]map[string]interface{})
			m["Member"] = make(map[string]interface{})
			m["Member"]["Status"] = "unparseable"
			return m, nil
		},
	}
	kvg := &FakeKVGetter{}
	c := &consulProvider{
		agent: s,
		kv:    kvg,
	}

	err := c.Healthcheck()

	assert.Error(t, err)
	s.AssertSelfCalledOnce(t)
}
