package env

import (
	"testing"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestMultiProviderGetMissing(t *testing.T) {
	p := MultiProvider{emptyConsul, missingConsul}

	value, err := p.Get(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetEmpty(t *testing.T) {
	p := MultiProvider{emptyConsul, emptyConsul}

	value, err := p.Get(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetSuccess(t *testing.T) {
	p := MultiProvider{emptyConsul, defaultConsul}

	value, err := p.Get(defaultKey)

	assert.Equal(t, value, string(defaultVal))
	assert.NoError(t, err)
}

func TestMultiProviderGetIntMissing(t *testing.T) {
	p := MultiProvider{emptyConsul, missingConsul}

	value, err := p.GetInt(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetIntInvalid(t *testing.T) {
	p := MultiProvider{emptyConsul, defaultConsul}

	value, err := p.GetInt(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetIntSuccess(t *testing.T) {
	p := MultiProvider{emptyConsul, intConsul}

	value, err := p.GetInt(defaultKey)

	assert.Equal(t, value, defaultInt)
	assert.NoError(t, err)
}

func TestMultiProviderGetBoolMissing(t *testing.T) {
	p := MultiProvider{emptyConsul, missingConsul}

	value, err := p.GetBool(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetBoolInvalid(t *testing.T) {
	p := MultiProvider{emptyConsul, defaultConsul}

	value, err := p.GetBool(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetBoolSuccess(t *testing.T) {
	p := MultiProvider{emptyConsul, boolConsul}

	value, err := p.GetBool(defaultKey)

	assert.True(t, value)
	assert.NoError(t, err)
}

func TestMultiProviderMonitor(t *testing.T) {
	i := uint64(0)
	expected1 := []byte("not empty")
	expected2 := []byte("also not empty")
	kv := &FakekvGetter{
		ListHook: func(string, *consul.QueryOptions) (consul.KVPairs, *consul.QueryMeta, error) {
			defer func() { i++ }()
			switch i {
			case 0:
				return consul.KVPairs{{Key: defaultKey, Value: []byte{}}}, &consul.QueryMeta{LastIndex: i + 1}, nil
			case 1:
				return consul.KVPairs{{Key: defaultKey, Value: expected1}}, &consul.QueryMeta{LastIndex: i + 1}, nil
			case 2:
				return consul.KVPairs{{Key: defaultKey, Value: expected2}}, &consul.QueryMeta{LastIndex: i + 1}, nil
			default:
				return nil, nil, merry.New("stop")
			}
		},
	}
	monitorConsul := &ConsulProvider{
		agent: NewFakeselferDefaultPanic(),
		kv:    kv,
	}

	p := MultiProvider{NewSystem(), monitorConsul}
	v := make(chan Value, 2)

	p.Monitor(defaultKey, v)
	val1 := <-v
	val2 := <-v

	assert.Equal(t, string(expected1), val1.Value)
	assert.Equal(t, string(expected2), val2.Value)
}
