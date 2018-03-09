package lb

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/sd"
)

func TestLeastNBalance(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testAddrs, nil
		},
	}
	rr := NewLeastN(res, 2)
	b, e := rr.Balance(testServiceName)
	if e != nil {
		t.Fatal(e)
	}

	assert.Contains(t, testAddrs, b)
}

func TestLeastNBalanceResolveError(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return nil, merry.New("some error")
		},
	}
	rr := NewLeastN(res, 2)

	node, e := rr.Balance(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, "", node)
}

func TestLeastNBalanceNoNodes(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return make([]string, 0), nil
		},
	}
	rr := NewLeastN(res, 2)

	node, e := rr.Balance(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, "", node)
}

func TestLeastNBalanceCalledTwice(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testAddrs, nil
		},
	}
	rr := NewLeastN(res, 2)

	node1, e := rr.Balance(testServiceName)
	node2, e := rr.Balance(testServiceName)

	assert.NoError(t, e)
	assert.NotEqual(t, node1, node2)
	assert.Contains(t, testAddrs, node1)
	assert.Contains(t, testAddrs, node2)
}

func TestLeastNBalanceLargeN(t *testing.T) {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testAddrs, nil
		},
	}
	rr := NewLeastN(res, 10)

	node1, e := rr.Balance(testServiceName)

	assert.NoError(t, e)
	assert.Contains(t, testAddrs, node1)
}
