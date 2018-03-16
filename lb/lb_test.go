package lb

import (
	"testing"
	"fmt"
)

const testServiceName = "testservice"

var testAddrs = []string{"10.0.0.1:9000", "10.0.0.2:9000", "10.0.0.3:9000", "10.0.0.4:9000"}

func setUpAddrs(i int) []string {
	s := make([]string, i)
	for i := range s {
		s[i] = fmt.Sprintf("10.0.0.%s", i)
	}
	return s
}

func benchmarkLB(lb Balancer, service string, b *testing.B) {
	for n := 0; n < b.N; n++ {
		lb.Balance(service)
	}
}
