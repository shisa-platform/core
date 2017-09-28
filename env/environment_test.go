package env

import (
	"testing"
	"os"

	"github.com/stretchr/testify/assert"
)

func setEnv(t *testing.T, envvar, value string) {
	err := os.Setenv(envvar, value)
	if err != nil {
		t.Fatalf("failed to set envvar")
	}
}

func TestGet(t *testing.T) {
	const envvar = "GO_SHISA_TEST_ENV_GET"

	value, ok := Get(envvar) // Should not exist.

	assert.Equal(t, value, "")
	assert.NotEqual(t, ok, nil)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "exists")

	value, ok = Get(envvar)

	assert.Equal(t, value, "exists")
	assert.Equal(t, ok, nil)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "")

	value, ok = Get(envvar)

	assert.Equal(t, value, "")
	assert.NotEqual(t, ok, nil)
}

func TestGetInt(t *testing.T) {
	const envvar = "GO_SHISA_TEST_ENV_GETINT"

	value, ok := GetInt(envvar) // Should not exist.

	assert.Equal(t, value, 0)
	assert.NotEqual(t, ok, nil)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "123")

	value, ok = GetInt(envvar)

	assert.Equal(t, value, 123)
	assert.Equal(t, ok, nil)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "false")

	value, ok = GetInt(envvar)

	assert.Equal(t, value, 0)
	assert.NotEqual(t, ok, nil)
}

func TestGetBool(t *testing.T) {
	const envvar = "GO_SHISA_TEST_ENV_GETBOOL"

	value, ok := GetBool(envvar) // Should not exist.

	assert.Equal(t, value, false)
	assert.NotEqual(t, ok, nil)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "true")

	value, ok = GetBool(envvar)

	assert.Equal(t, value, true)
	assert.Equal(t, ok, nil)

	defer os.Unsetenv(envvar)
	setEnv(t, envvar, "false")

	value, ok = GetBool(envvar)

	assert.Equal(t, value, false)
	assert.Equal(t, ok, nil)
}
