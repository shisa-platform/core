package metrics

import (
	"encoding/json"
	"sync"
)

type Measurer interface {
	Measure(key string, elapsed float64)
}

type ConcurrentMetrics struct {
	mux    sync.RWMutex
	values map[string]*Timing
}

func New() *ConcurrentMetrics {
	return &ConcurrentMetrics{
		values: make(map[string]*Timing),
	}
}

func (m *ConcurrentMetrics) Measure(key string, elapsed float64) {
	m.mux.Lock()
	defer m.mux.Unlock()
	value, found := m.values[key]
	if !found {
		value = new(Timing)
		m.values[key] = value
	}
	value.Accum(elapsed)
}

func (m *ConcurrentMetrics) Get(key string) (*Timing, bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	value, ok := m.values[key]
	return value, ok
}

func (m *ConcurrentMetrics) Do(f func(string, *Timing)) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	for key, value := range m.values {
		f(key, value)
	}
}

func (m *ConcurrentMetrics) MarshalJSON() ([]byte, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return json.Marshal(m.values)
}
