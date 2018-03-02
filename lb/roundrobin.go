package lb

import (
	"sync/atomic"

	"github.com/ansel1/merry"
)

var _ Balancer = &roundRobin{}

func NewRoundRobin() Balancer {
	return &roundRobin{
		c: 0,
	}
}

type roundRobin struct {
	c uint64
}

func (rr *roundRobin) Balance(nodes []string) (string, merry.Error) {
	if len(nodes) <= 0 {
		return "", merry.New("no nodes available")
	}
	old := atomic.AddUint64(&rr.c, 1) - 1
	idx := old % uint64(len(nodes))

	return nodes[idx], nil
}
