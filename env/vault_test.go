package env

import (
	"testing"

	"github.com/ansel1/merry"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

const (
	defaultPrefix = "zalgo"
)

var (
	missingVault = newFakeVault([]byte(""), merry.New("missing"))
	emptyVault   = newFakeVault([]byte(""), nil)
	defaultVault = newFakeVault(defaultVal, nil)
	intVault     = newFakeVault(defaultIntVal, nil)
	boolVault    = newFakeVault(defaultBoolVal, nil)
)

func newFakeVault(val []byte, err error) *VaultProvider {
	return &VaultProvider{
		logical: &FakelogicalReader{
			ReadHook: func(s string) (*vaultapi.Secret, error) {
				return &vaultapi.Secret{}, nil
			},
		},
	}
}

func TestNewVault(t *testing.T) {
	ac := &vaultapi.Client{}
	prefix := "secret/"

	c := NewVault(ac, prefix)

	assert.NotNil(t, c)
	assert.NotNil(t, c.logical)
	assert.Equal(t, prefix, c.prefix)
}

func TestVaultProviderGet(t *testing.T) {
	data := map[string]interface{}{string(defaultKey): string(defaultVal)}
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return &vaultapi.Secret{Data: data}, nil
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.Get(defaultKey)

	assert.Equal(t, string(defaultVal), r)
	assert.NoError(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetError(t *testing.T) {
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return nil, merry.New("error reading secret")
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetEmpty(t *testing.T) {
	data := map[string]interface{}{string(defaultKey): ""}
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return &vaultapi.Secret{Data: data}, nil
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetMissing(t *testing.T) {
	data := make(map[string]interface{})
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return &vaultapi.Secret{Data: data}, nil
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetNil(t *testing.T) {
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return nil, nil
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetBadString(t *testing.T) {
	data := map[string]interface{}{string(defaultKey): vaultapi.Secret{}}
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return &vaultapi.Secret{Data: data}, nil
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.Get(defaultKey)

	assert.Equal(t, "", r)
	assert.Error(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetInt(t *testing.T) {
	data := map[string]interface{}{string(defaultKey): "1"}
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return &vaultapi.Secret{Data: data}, nil
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.GetInt(defaultKey)

	assert.Equal(t, defaultInt, r)
	assert.NoError(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetIntError(t *testing.T) {
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return nil, merry.New("error reading secret")
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.GetInt(defaultKey)

	assert.Equal(t, 0, r)
	assert.Error(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetIntParseFailure(t *testing.T) {
	data := map[string]interface{}{string(defaultKey): "NaN"}
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return &vaultapi.Secret{Data: data}, nil
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.GetInt(defaultKey)

	assert.Equal(t, 0, r)
	assert.Error(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetBool(t *testing.T) {
	data := map[string]interface{}{string(defaultKey): "true"}
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return &vaultapi.Secret{Data: data}, nil
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.GetBool(defaultKey)

	assert.Equal(t, true, r)
	assert.NoError(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetBoolError(t *testing.T) {
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return nil, merry.New("error reading secret")
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.GetBool(defaultKey)

	assert.Equal(t, false, r)
	assert.Error(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}

func TestVaultProviderGetBoolParseFailure(t *testing.T) {
	data := map[string]interface{}{string(defaultKey): "not a bool"}
	reader := &FakelogicalReader{
		ReadHook: func(s string) (*vaultapi.Secret, error) {
			return &vaultapi.Secret{Data: data}, nil
		},
	}
	vaultProvider := &VaultProvider{
		logical: reader,
		prefix:  defaultPrefix,
	}

	r, err := vaultProvider.GetBool(defaultKey)

	assert.Equal(t, false, r)
	assert.Error(t, err)
	reader.AssertReadCalledOnceWith(t, defaultPrefix+defaultKey)
}
