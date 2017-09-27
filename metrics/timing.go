package metrics

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
)

type Timing struct {
	mux   sync.RWMutex
	count int64
	total float64
}

func (t *Timing) Accum(delta float64) {
	t.mux.Lock()
	t.total += delta
	t.count++
	t.mux.Unlock()
}

func (t *Timing) Time() float64 {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return t.total
}

func (t *Timing) Count() int64 {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return t.count
}

func (t *Timing) Avg() float64 {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return t.calcuateAverage()
}

func (t *Timing) calcuateAverage() float64 {
	if t.count == 0 {
		return 0.0
	}
	return t.total / float64(t.count)
}

func (t *Timing) String() string {
	return t.serialize().String()
}

func (t *Timing) MarshalJSON() ([]byte, error) {
	return t.serialize().Bytes(), nil
}

func (t *Timing) serialize() *bytes.Buffer {
	var buf bytes.Buffer
	t.mux.RLock()
	defer t.mux.RUnlock()
	fmt.Fprintf(&buf,
		"{\"avg\": %s, \"count\": %d, \"total\": %s}",
		strconv.FormatFloat(t.calcuateAverage(), 'g', -1, 64),
		t.count,
		strconv.FormatFloat(t.total, 'g', -1, 64))

	return &buf
}
