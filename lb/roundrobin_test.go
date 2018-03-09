package lb

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/percolate/shisa/sd"
	"github.com/stretchr/testify/assert"
)

func TestRoundRobinBalance(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testAddrs, nil
		},
	}
	rr := NewRoundRobin(res)
	results := make([]string, len(testAddrs))

	var e error
	for i := range testAddrs {
		results[i], e = rr.Balance(testServiceName)
		if e != nil {
			t.Fatal(e)
		}
	}

	assert.ElementsMatch(t, testAddrs, results)
}

func TestRoundRobinBalanceResolveError(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return nil, merry.New("some error")
		},
	}
	rr := NewRoundRobin(res)

	node, e := rr.Balance(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, "", node)
}

func TestRoundRobinBalanceNoNodes(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return make([]string, 0), nil
		},
	}
	rr := NewRoundRobin(res)

	node, e := rr.Balance(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, "", node)
}
