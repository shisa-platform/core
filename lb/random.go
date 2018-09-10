package lb

import (
	"math/rand"
	"time"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/sd"
)

var _ Balancer = &randomLB{}

type randomLB struct {
	res sd.Resolver
}

// NewRandom returns a Balancer that randomly selects one of the current nodes
// supplied by its Resolver
func NewRandom(res sd.Resolver) Balancer {
	return &randomLB{
		res: res,
	}
}

func (r *randomLB) Balance(service string) (string, merry.Error) {
	nodes, err := r.res.Resolve(service)
	if err != nil {
		return "", err.Prepend("random balancer: balance")
	}

	switch len(nodes) {
	case 0:
		return "", merry.New("random balancer: balance: no nodes available")
	case 1:
		return nodes[0], nil
	}

	rando := rand.New(rand.NewSource(time.Now().Unix()))

	return nodes[rando.Intn(len(nodes))], nil
}
