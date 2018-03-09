package lb

import (
	"github.com/ansel1/merry"
)

// Balancer defines a Balance method that returns an addr string and
// merry.Error for a service node given a string denoting the service name
type Balancer interface {
	Balance(service string) (addr string, err merry.Error)
}
