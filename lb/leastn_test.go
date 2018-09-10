package lb

import (
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/sd"
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
	cache := &leastConnsCache{
		n: 2,
		r: make(map[string]*nodeGroup),
	}
	rr := &CacheBalancer{
		Cache:    cache,
		Resolver: res,
	}

	node1, e := rr.Balance(testServiceName)
	conns1 := cache.connsForService(testServiceName)
	node2, e := rr.Balance(testServiceName)
	conns2 := cache.connsForService(testServiceName)

	assert.NoError(t, e)
	assert.Contains(t, testAddrs, node1)
	assert.Contains(t, testAddrs, node2)
	assert.Equal(t, uint64(1), conns1)
	assert.Equal(t, uint64(2), conns2)
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

func TestMergeNodes(t *testing.T) {
	newGroup := newNodeGroup(testAddrs)
	before := make([]*node, len(newGroup.nodes))
	copy(before, newGroup.nodes)

	old := newNodeGroup([]string{"10.0.0.4:9000", "10.0.0.5:9000", "10.0.0.6:9000"})

	old.merge(newGroup.nodes)
	result := old.nodes
	resAddrs := make([]string, len(result))

	for i, r := range result {
		resAddrs[i] = r.addr
	}

	assert.Len(t, result, len(before))
	assert.Subset(t, result, before)
	assert.Contains(t, resAddrs, "10.0.0.4:9000")
	assert.NotContains(t, resAddrs, "10.0.0.5:9000")
	assert.NotContains(t, resAddrs, "10.0.0.6:9000")
}

func TestNewNodeGroup(t *testing.T) {
	result := newNodeGroup(testAddrs)

	resAddrs := make([]string, len(result.nodes))
	for i, r := range result.nodes {
		resAddrs[i] = r.addr
	}

	assert.Equal(t, len(testAddrs), len(result.nodes))
	assert.ElementsMatch(t, testAddrs, resAddrs)
}

func setUpLeastN(i int, testAddrs []string) Balancer {
	res := &sd.FakeResolver{
		ResolveHook: func(name string) ([]string, merry.Error) {
			return testAddrs, nil
		},
	}
	return NewLeastN(res, i)
}

func Benchmark2Least10Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(2, setUpAddrs(10)), "testservice", b)
}
func Benchmark2Least100Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(2, setUpAddrs(100)), "testservice", b)
}
func Benchmark2Least1000Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(2, setUpAddrs(1000)), "testservice", b)
}
func Benchmark2Least10000Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(2, setUpAddrs(10000)), "testservice", b)
}
func Benchmark3Least10Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(3, setUpAddrs(10)), "testservice", b)
}
func Benchmark3Least100Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(3, setUpAddrs(100)), "testservice", b)
}
func Benchmark3Least1000Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(3, setUpAddrs(1000)), "testservice", b)
}
func Benchmark3Least10000Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(3, setUpAddrs(10000)), "testservice", b)
}
func Benchmark4Least10Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(4, setUpAddrs(10)), "testservice", b)
}
func Benchmark4Least100Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(4, setUpAddrs(100)), "testservice", b)
}
func Benchmark4Least1000Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(4, setUpAddrs(1000)), "testservice", b)
}
func Benchmark4Least10000Nodes(b *testing.B) {
	benchmarkLB(setUpLeastN(4, setUpAddrs(10000)), "testservice", b)
}
