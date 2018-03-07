package lb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testServiceName = "testservice"

var testHosts = []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"}

func TestMergeNodes(t *testing.T) {
	one := NewNodeGroup(testHosts).nodes
	old := NewNodeGroup([]string{"10.0.0.4", "10.0.0.5", "10.0.0.6"}).nodes

	result := UpdateNodes(one, old)
	resHosts := make([]string, len(result))

	for i, r := range result {
		resHosts[i] = r.host
	}

	assert.Len(t, result, len(one))
	assert.Subset(t, result, one)
	assert.Contains(t, resHosts, "10.0.0.4")
	assert.NotContains(t, resHosts, "10.0.0.5")
	assert.NotContains(t, resHosts, "10.0.0.6")
}

func TestNewNodeGroup(t *testing.T) {
	result := NewNodeGroup(testHosts)

	resHosts := make([]string, len(result.nodes))
	for i, r := range result.nodes {
		resHosts[i] = r.host
	}

	assert.Equal(t, len(testHosts), len(result.nodes))
	assert.ElementsMatch(t, testHosts, resHosts)
}
