package env

import (
	"testing"

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
