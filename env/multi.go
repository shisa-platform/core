package env

import (
	"github.com/ansel1/merry"
	"go.uber.org/multierr"
)

var _ Provider = (MultiProvider)(nil)

type MultiProvider []Provider

func (p MultiProvider) Get(name string) (string, merry.Error) {
	var err error
	for _, pro := range p {
		val, perr := pro.Get(name)
		if perr == nil {
			return val, nil
		}
		err = multierr.Combine(err, perr)
	}
	return "", merry.Prependf(err, "multi env provider: get %q", name)
}

func (p MultiProvider) GetInt(name string) (int, merry.Error) {
	var err error
	for _, pro := range p {
		val, perr := pro.GetInt(name)
		if perr == nil {
			return val, nil
		}
		err = multierr.Combine(err, perr)
	}
	return 0, merry.Prependf(err, "multi env provider: get int %q", name)
}

func (p MultiProvider) GetBool(name string) (bool, merry.Error) {
	var err error
	for _, pro := range p {
		val, perr := pro.GetBool(name)
		if perr == nil {
			return val, nil
		}
		err = multierr.Combine(err, perr)
	}
	return false, merry.Prependf(err, "multi env provider: get bool %q", name)
}

func (p MultiProvider) Monitor(name string, v chan<- Value) {
	for _, pro := range p {
		pro.Monitor(name, v)
	}
}
