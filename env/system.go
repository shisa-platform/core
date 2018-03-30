package env

import (
	"os"
	"strconv"

	"github.com/ansel1/merry"
)

var (
	DefaultProvider = NewSystem()
)

type systemProvider struct{}

func (p *systemProvider) Get(name string) (string, merry.Error) {
	if value, ok := os.LookupEnv(name); ok {
		if value == "" {
			return "", NameEmpty.Prepend("system env provider: get").Append(name)
		}
		return value, nil
	}

	return "", NameNotSet.Prepend("system env provider: get").Append(name)
}

func (p *systemProvider) GetInt(name string) (int, merry.Error) {
	if value, ok := os.LookupEnv(name); ok {
		if r, err := strconv.Atoi(value); err == nil {
			return r, nil
		} else {
			return 0, merry.Prepend(err, "system env provider: get int").Append(name)
		}
	}

	return 0, NameNotSet.Prepend("system env provider: get int").Append(name)
}

func (p *systemProvider) GetBool(name string) (bool, merry.Error) {
	if value, ok := os.LookupEnv(name); ok {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return false, merry.Prepend(err, "system env provider: get bool").Append(name)
		}

		return b, nil
	}

	return false, NameNotSet.Prepend("system env provider: get bool").Append(name)
}

func (p *systemProvider) Monitor(string, chan<- Value) {
	// N.B. do nothing, system enviroment vars are not dynamic
}

func NewSystem() Provider {
	return new(systemProvider)
}

func Get(name string) (string, merry.Error) {
	return DefaultProvider.Get(name)
}

func GetInt(name string) (int, merry.Error) {
	return DefaultProvider.GetInt(name)
}

func GetBool(name string) (bool, merry.Error) {
	return DefaultProvider.GetBool(name)
}
