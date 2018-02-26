package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testServiceName = "testservice"

var testHosts = []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"}

func TestCache(t *testing.T) {
	c := NewCache()
	c.Update(testServiceName, testHosts)

	res := c.Get(testServiceName)

	assert.ElementsMatch(t, testHosts, res)
}

func TestCacheOrder(t *testing.T) {
	c := NewCache()
	c.Update(testServiceName, testHosts)

	res1 := c.Get(testServiceName)

	l := len(res1)
	rev := make([]string, l)
	for i, x := range res1 {
		rev[l-i-1] = x
	}
	c.Update(testServiceName, rev)

	res2 := c.Get(testServiceName)

	assert.ElementsMatch(t, testHosts, res1, "original")
	assert.ElementsMatch(t, testHosts, res2, "reversed")
	assert.Equal(t, res1, res2)
}
