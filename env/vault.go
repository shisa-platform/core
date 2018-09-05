package env

import (
	"github.com/ansel1/merry"
	vault "github.com/hashicorp/vault/api"
	"strconv"
)

var _ Provider = (*VaultProvider)(nil)
var _ logicalReader = (*vault.Logical)(nil)

//go:generate charlatan -output=./vaultlogicalreader_charlatan.go logicalReader

type logicalReader interface {
	Read(path string) (*vault.Secret, error)
}

type VaultProvider struct {
	logical logicalReader

	prefix string
}

func NewVault(c *vault.Client, prefix string) *VaultProvider {
	return &VaultProvider{
		logical: c.Logical(),
		prefix:  prefix,
	}
}

func (v *VaultProvider) Get(name string) (string, merry.Error) {
	secret, err := v.logical.Read(v.prefix+name)
	if err != nil {
		return "", merry.Prepend(err, "vault env provider: get").Append(name)
	}
	if secret == nil {
		return "", NameNotSet.Prepend("vault env provider: get").Append(name)
	}

	rawVal, ok := secret.Data[name]
	if !ok {
		return "", NameNotSet.Prepend("vault env provider: get").Append(name)
	}

	r, ok := rawVal.(string)
	if !ok {
		return "", NameNotSet.Prepend("vault env provider: get").Append(name)
	}

	if r == "" {
		return "", NameEmpty.Prepend("vault env provider: get").Append(name)
	}

	return r, nil
}

func (v *VaultProvider) GetInt(name string) (int, merry.Error) {
	stringVal, merr := v.Get(name)
	if merr != nil {
		return 0, merr
	}

	r, err := strconv.Atoi(stringVal)
	if err != nil {
		return 0, merry.Prepend(err, "vault env provider: get int").Append(name)
	}

	return r, nil
}

func (v *VaultProvider) GetBool(name string) (bool, merry.Error) {
	stringVal, merr := v.Get(name)
	if merr != nil {
		return false, merr
	}

	r, err := strconv.ParseBool(stringVal)
	if err != nil {
		return false, merry.Prepend(err, "vault env provider: get bool").Append(name)
	}

	return r, nil
}

func (v *VaultProvider) Monitor(name string, ch chan<- Value) {
	// N.B. do nothing, vault secrets are not able to be monitored
}
