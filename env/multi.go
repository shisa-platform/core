package env

import (
	"github.com/ansel1/merry"
	"go.uber.org/multierr"
)

var _ Provider = (*MultiProvider)(nil)

type MultiProvider struct {
	providers []Provider
}

func NewMulti(providers ...Provider) Provider {
	return &MultiProvider{providers}
}

func (p *MultiProvider) Get(name string) (string, merry.Error) {
	var err error
	for _, pro := range p.providers {
		val, perr := pro.Get(name)
		if perr == nil {
			return val, nil
		}
		err = multierr.Combine(err, perr)
	}
	return "", merry.Prepend(err, "multi env provider: get").Append(name)
}

func (p *MultiProvider) GetInt(name string) (int, merry.Error) {
	var err error
	for _, pro := range p.providers {
		val, perr := pro.GetInt(name)
		if perr == nil {
			return val, nil
		}
		err = multierr.Combine(err, perr)
	}
	return 0, merry.Prepend(err, "multi env provider: get int").Append(name)
}

func (p *MultiProvider) GetBool(name string) (bool, merry.Error) {
	var err error
	for _, pro := range p.providers {
		val, perr := pro.GetBool(name)
		if perr == nil {
			return val, nil
		}
		err = multierr.Combine(err, perr)
	}
	return false, merry.Prepend(err, "multi env provider: get bool").Append(name)
}

func (*MultiProvider) Monitor(string, chan<- Value) {
	// N.B. do nothing, multi provider enviroment vars are not dynamic
}
