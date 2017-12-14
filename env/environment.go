package env

import (
	"os"
	"strconv"

	"github.com/ansel1/merry"
)

var (
	defaultProvider = NewSystem()
)

type systemProvider struct{}

func (p *systemProvider) Get(name string) (string, merry.Error) {
	if value, ok := os.LookupEnv(name); ok {
		if value == "" {
			return "", NameEmpty.Here().WithValue("name", name)
		}
		return value, nil
	}

	return "", NameNotSet.Here().WithValue("name", name)
}

func (p *systemProvider) GetInt(name string) (int, merry.Error) {
	if value, ok := os.LookupEnv(name); ok {
		if r, err := strconv.Atoi(value); err == nil {
			return r, nil
		} else {
			return 0, merry.Wrap(err).WithValue("name", name)
		}
	}

	return 0, NameNotSet.Here().WithValue("name", name)
}

func (p *systemProvider) GetBool(name string) (bool, merry.Error) {
	if value, ok := os.LookupEnv(name); ok {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return false, merry.Wrap(err).WithValue("name", name)
		}

		return b, nil
	}

	return false, NameNotSet.Here().WithValue("name", name)
}

func (p *systemProvider) Monitor(string, <-chan Value) {
	// N.B. do nothing, system enviroment vars are not dynamic
}

func NewSystem() Provider {
	return new(systemProvider)
}

func Get(name string) (string, merry.Error) {
	return defaultProvider.Get(name)
}

func GetInt(name string) (int, merry.Error) {
	return defaultProvider.GetInt(name)
}

func GetBool(name string) (bool, merry.Error) {
	return defaultProvider.GetBool(name)
}
