package env

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
)

var (
	missingProvider = newFakeProvider("", NameNotSet)
	emptyProvider   = newFakeProvider("", NameEmpty)
	defaultProvider = newFakeProvider(defaultVal, nil)
	intProvider     = &FakeProvider{
		GetIntHook: func(string) (int, merry.Error) {
			return defaultInt, nil
		},
	}
	boolProvider = &FakeProvider{
		GetBoolHook: func(string) (bool, merry.Error) {
			return defaultBool, nil
		},
	}
)

func newFakeProvider(val string, err merry.Error) Provider {
	return &FakeProvider{
		GetHook: func(s string) (string, merry.Error) {
			if err != nil {
				return "", err
			}

			return val, nil
		},
		GetIntHook: func(string) (int, merry.Error) {
			if err != nil {
				return 0, err
			}

			return defaultInt, nil
		},
		GetBoolHook: func(string) (bool, merry.Error) {
			if err != nil {
				return false, err
			}

			return defaultBool, nil
		},
	}
}

func TestMultiProviderGetMissing(t *testing.T) {
	p := MultiProvider{emptyProvider, missingProvider}

	value, err := p.Get(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetEmpty(t *testing.T) {
	p := MultiProvider{emptyProvider, emptyProvider}

	value, err := p.Get(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGet(t *testing.T) {
	p := MultiProvider{emptyProvider, defaultProvider}

	value, err := p.Get(defaultKey)

	assert.Equal(t, value, string(defaultVal))
	assert.NoError(t, err)
}

func TestMultiProviderGetIntMissing(t *testing.T) {
	errorIntProvider := &FakeProvider{
		GetIntHook: func(string) (int, merry.Error) {
			return 0, NameNotSet
		},
	}
	p := MultiProvider{emptyProvider, errorIntProvider}

	value, err := p.GetInt(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetIntInvalid(t *testing.T) {
	errorIntProvider := &FakeProvider{
		GetIntHook: func(string) (int, merry.Error) {
			return 0, merry.New("not an integer")
		},
	}
	p := MultiProvider{emptyProvider, errorIntProvider}

	value, err := p.GetInt(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetInt(t *testing.T) {
	p := MultiProvider{emptyProvider, intProvider}

	value, err := p.GetInt(defaultKey)

	assert.Equal(t, value, defaultInt)
	assert.NoError(t, err)
}

func TestMultiProviderGetBoolMissing(t *testing.T) {
	errorBoolProvider := &FakeProvider{
		GetBoolHook: func(string) (bool, merry.Error) {
			return false, NameNotSet
		},
	}
	p := MultiProvider{emptyProvider, errorBoolProvider}

	value, err := p.GetBool(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetBoolInvalid(t *testing.T) {
	errorBoolProvider := &FakeProvider{
		GetBoolHook: func(string) (bool, merry.Error) {
			return false, merry.New("not a boolean")
		},
	}
	p := MultiProvider{emptyProvider, errorBoolProvider}

	value, err := p.GetBool(defaultKey)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestMultiProviderGetBool(t *testing.T) {
	p := MultiProvider{emptyProvider, boolProvider}

	value, err := p.GetBool(defaultKey)

	assert.True(t, value)
	assert.NoError(t, err)
}
