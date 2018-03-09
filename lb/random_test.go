package lb

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/percolate/shisa/sd"
	"github.com/stretchr/testify/assert"
)

func TestRandomBalance(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testAddrs, nil
		},
	}
	rr := NewRandom(res)

	result, e := rr.Balance(testServiceName)
	if e != nil {
		t.Fatal(e)
	}

	assert.Contains(t, testAddrs, result)
}

func TestRandomBalanceResolveError(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return nil, merry.New("some error")
		},
	}
	rr := NewRandom(res)

	node, e := rr.Balance(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, "", node)
}

func TestRandomBalanceNoNodes(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return make([]string, 0), nil
		},
	}
	rr := NewRandom(res)

	node, e := rr.Balance(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, "", node)
}
