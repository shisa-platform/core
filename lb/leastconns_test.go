package lb

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/percolate/shisa/sd"
	"github.com/stretchr/testify/assert"
)

func TestLeastConnsBalance(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testHosts, nil
		},
	}
	rr := NewLeastConns(res, 2)
	b, e := rr.Balance(testServiceName)
	if e != nil {
		t.Fatal(e)
	}

	assert.Contains(t, testHosts, b)
}

func TestLeastConnsBalanceResolveError(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return nil, merry.New("some error")
		},
	}
	rr := NewLeastConns(res, 2)

	node, e := rr.Balance(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, "", node)
}

func TestLeastConnsBalanceNoNodes(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return make([]string, 0), nil
		},
	}
	rr := NewLeastConns(res, 2)

	node, e := rr.Balance(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, "", node)
}

func TestLeastConnsBalanceCalledTwice(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testHosts, nil
		},
	}
	rr := NewLeastConns(res, 2)

	node1, e := rr.Balance(testServiceName)
	node2, e := rr.Balance(testServiceName)

	assert.NoError(t, e)
	assert.NotEqual(t, node1, node2)
	assert.Contains(t, testHosts, node1)
	assert.Contains(t, testHosts, node2)
}

func TestLeastConnsBalanceLargeN(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testHosts, nil
		},
	}
	rr := NewLeastConns(res, 10)

	node1, e := rr.Balance(testServiceName)

	assert.NoError(t, e)
	assert.Contains(t, testHosts, node1)
}
