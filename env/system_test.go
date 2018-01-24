package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setEnv(t *testing.T, name, value string) {
	t.Helper()
	err := os.Setenv(name, value)
	if err != nil {
		t.Fatalf("failed to set envvar %q: %q - %v", name, value, err)
	}
}

func TestGet(t *testing.T) {
	const envvar = "GO_SHISA_TEST_ENV_GET"

	value, err := Get(envvar) // Should not exist.

	assert.Empty(t, value)
	assert.Error(t, err)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "exists")

	value, err = Get(envvar)

	assert.Equal(t, value, "exists")
	assert.NoError(t, err)

	setEnv(t, envvar, "")

	value, err = Get(envvar)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestGetInt(t *testing.T) {
	const envvar = "GO_SHISA_TEST_ENV_GETINT"

	value, err := GetInt(envvar) // Should not exist.

	assert.Empty(t, value)
	assert.Error(t, err)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "123")

	value, err = GetInt(envvar)

	assert.Equal(t, value, 123)
	assert.NoError(t, err)

	setEnv(t, envvar, "false")

	value, err = GetInt(envvar)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestGetBool(t *testing.T) {
	const envvar = "GO_SHISA_TEST_ENV_GETBOOL"

	value, err := GetBool(envvar) // Should not exist.

	assert.Empty(t, value)
	assert.Error(t, err)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "true")

	value, err = GetBool(envvar)

	assert.True(t, value)
	assert.NoError(t, err)

	setEnv(t, envvar, "false")

	value, err = GetBool(envvar)

	assert.False(t, value)
	assert.NoError(t, err)

	setEnv(t, envvar, "zuul")

	value, err = GetBool(envvar)

	assert.Empty(t, value)
	assert.Error(t, err)
}

func TestHealthcheck(t *testing.T) {
	cut := NewSystem()
	assert.NotNil(t, cut)

	err := cut.Healthcheck()
	assert.NoError(t, err)
}

func TestMonitor(t *testing.T) {
	cut := NewSystem()
	assert.NotNil(t, cut)

	const envvar = "GO_SHISA_TEST_MONITOR"

	value, err := cut.Get(envvar) // Should not exist.

	assert.Empty(t, value)
	assert.Error(t, err)

	ch := make(chan Value, 1)
	cut.Monitor(envvar, ch)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "exists")

	value, err = cut.Get(envvar)

	assert.Equal(t, value, "exists")
	assert.NoError(t, err)

	assert.Empty(t, ch)
}
