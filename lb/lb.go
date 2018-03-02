package lb

import "github.com/ansel1/merry"

type Balancer interface {
	Balance([]string) (string, merry.Error)
}
