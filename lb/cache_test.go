package lb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRRCache(t *testing.T) {
	c := newRRCache()

	res := c.Next(testServiceName, testAddrs)

	assert.Contains(t, testAddrs[0], res)
}

func TestRRCacheOrder(t *testing.T) {
	c := newRRCache()

	res1 := c.Next(testServiceName, testAddrs)

	l := len(testAddrs)
	rev := make([]string, l)
	for i, x := range testAddrs {
		rev[l-i-1] = x
	}

	res2 := c.Next(testServiceName, rev)

	assert.Equal(t, testAddrs[0], res1)
	assert.Equal(t, testAddrs[1], res2)
}

func TestRRCacheAdditon(t *testing.T) {
	c := newRRCache()

	res1 := c.Next(testServiceName, testAddrs)

	l := len(testAddrs)
	rev := make([]string, l+1)
	for i, x := range testAddrs {
		rev[l-i-1] = x
	}
	rev[0] = "10.0.0.5"

	res2 := c.Next(testServiceName, rev)

	assert.Equal(t, testAddrs[0], res1)
	assert.Equal(t, testAddrs[1], res2)
}

func TestMergeNodes(t *testing.T) {
	one := newNodeGroup(testAddrs).nodes
	old := newNodeGroup([]string{"10.0.0.4:9000", "10.0.0.5:9000", "10.0.0.6:9000"}).nodes

	result := updateNodes(one, old)
	resAddrs := make([]string, len(result))

	for i, r := range result {
		resAddrs[i] = r.addr
	}

	assert.Len(t, result, len(one))
	assert.Subset(t, result, one)
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
