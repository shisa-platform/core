package lb

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/shisa-platform/core/sd"
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

func setUpRoundRobin(testAddrs []string) Balancer {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testAddrs, nil
		},
	}
	return NewRoundRobin(res)
}

func BenchmarkRoundRobin10Nodes(b *testing.B) {
	benchmarkLB(setUpRoundRobin(setUpAddrs(10)), "testservice", b)
}
func BenchmarkRoundRobin100Nodes(b *testing.B) {
	benchmarkLB(setUpRoundRobin(setUpAddrs(100)), "testservice", b)
}
func BenchmarkRoundRobin1000Nodes(b *testing.B) {
	benchmarkLB(setUpRoundRobin(setUpAddrs(1000)), "testservice", b)
}
func BenchmarkRoundRobin10000Nodes(b *testing.B) {
	benchmarkLB(setUpRoundRobin(setUpAddrs(10000)), "testservice", b)
}
