package context

import (
	stdctx "context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	cut := Get(stdctx.Background())
	defer Put(cut)

	assert.NotNil(t, cut)
	assert.Equal(t, "", cut.RequestID())
	assert.Nil(t, cut.Actor())
}
