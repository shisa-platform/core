package metrics

import (
	"bytes"
	"fmt"
)

func (m *ConcurrentMetrics) String() string {
	m.mux.RLock()
	defer m.mux.RUnlock()
	var b bytes.Buffer
	fmt.Fprintf(&b, "{")
	first := true
	for key, value := range m.values {
		if !first {
			fmt.Fprintf(&b, ", ")
		}
		fmt.Fprintf(&b, "%q: %v", key, value)
		first = false
	}
	fmt.Fprintf(&b, "}")

	return b.String()
}
