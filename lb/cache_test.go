package lb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	c := NewRRCache()

	res := c.Next(testServiceName, testHosts)

	assert.Contains(t, testHosts[0], res)
}

func TestCacheOrder(t *testing.T) {
	c := NewRRCache()

	res1 := c.Next(testServiceName, testHosts)

	l := len(testHosts)
	rev := make([]string, l)
	for i, x := range testHosts {
		rev[l-i-1] = x
	}

	res2 := c.Next(testServiceName, rev)

	assert.Equal(t, testHosts[0], res1)
	assert.Equal(t, testHosts[1], res2)
}

func TestCacheAdditon(t *testing.T) {
	c := NewRRCache()

	res1 := c.Next(testServiceName, testHosts)

	l := len(testHosts)
	rev := make([]string, l+1)
	for i, x := range testHosts {
		rev[l-i-1] = x
	}
	rev[0] = "10.0.0.5"

	res2 := c.Next(testServiceName, rev)

	assert.Equal(t, testHosts[0], res1)
	assert.Equal(t, testHosts[1], res2)
}
