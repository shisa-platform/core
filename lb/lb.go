package lb

import (
	"github.com/ansel1/merry"
)

// Balancer is a load balancer that wraps a service discovery
// instance to distribute access across the nodes of a service.
type Balancer interface {
	// Balance will return an address for a service or an empty
	// string if no nodes are available.
	Balance(service string) (addr string, err merry.Error)
}
