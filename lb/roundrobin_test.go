package lb

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/percolate/shisa/sd"
	"github.com/stretchr/testify/assert"
)

const testServiceName = "testservice"

var testHosts = []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"}

func TestRoundRobinBalance(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testHosts, nil
		},
	}
	rr := NewRoundRobin(res)
	results := make([]string, len(testHosts))

	var e error
	for i := range testHosts {
		results[i], e = rr.Balance(testServiceName)
		if e != nil {
			t.Fatal(e)
		}
	}

	assert.ElementsMatch(t, testHosts, results)
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
