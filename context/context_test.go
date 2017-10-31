package context

import (
	"context"
	"testing"
	"time"

	"errors"
	"github.com/percolate/shisa/context/contexttest"
	"github.com/percolate/shisa/models/modelstest"
	"github.com/stretchr/testify/assert"
)

var _ context.Context = &contexttest.FakeContext{}

func makeContext(requestID string) *Context {
	actor := &modelstest.FakeUser{}
	parent := &contexttest.FakeContext{}

	c := New(parent, requestID, actor)

	return c
}

func getContextForParent(parent context.Context) *Context {
	actor := &modelstest.FakeUser{}

	c := New(parent, "999999", actor)

	return c
}

func TestNew(t *testing.T) {
	actor := &modelstest.FakeUser{}
	parent := &contexttest.FakeContext{}
	id := "555"

	c := New(parent, id, actor)

	assert.Equal(t, parent, c.Context)
	assert.Equal(t, id, c.RequestID)
	assert.Equal(t, actor, c.Actor)
}

func TestDeadline(t *testing.T) {
	deadline := time.Time{}
	ok := false

	parent := &contexttest.FakeContext{
		DeadlineHook: func() (deadline time.Time, ok bool) {
			return deadline, ok
		},
	}
	c := getContextForParent(parent)

	d, y := c.Deadline()

	parent.AssertDeadlineCalledOnce(t)

	assert.Equal(t, d, deadline)
	assert.Equal(t, y, ok)

}

func TestDone(t *testing.T) {
	channelval := struct{}{}
	parent := &contexttest.FakeContext{
		DoneHook: func() (ret0 <-chan struct{}) {
			d := make(chan struct{})
			go func() {
				d <- channelval
				close(d)
			}()
			return d
		},
	}
	c := getContextForParent(parent)

	result := <-c.Done()

	parent.AssertDoneCalledOnce(t)
	assert.Equal(t, result, channelval)
}

func TestErr(t *testing.T) {
	err := errors.New("New Error")
	parent := &contexttest.FakeContext{
		ErrHook: func() (ret0 error) {
			return err
		},
	}
	c := getContextForParent(parent)

	result := c.Err()

	parent.AssertErrCalledOnce(t)
	assert.Equal(t, result, err)
}

func TestValueID(t *testing.T) {
	pkey := "ParentKey"
	pval := true
	parent := &contexttest.FakeContext{
		ValueHook: func(key interface{}) (ret0 interface{}) {
			return pval
		},
	}

	c := getContextForParent(parent)

	idVal := c.Value(IDKey)
	actorVal := c.Value(ActorKey)
	parentVal := c.Value(pkey)

	assert.Equal(t, idVal, c.RequestID)
	assert.Equal(t, actorVal, c.Actor)

	parent.AssertValueOnceCalledWith(t, pkey)

	calledVal, found := parent.ValueResultsForCall(pkey)
	assert.True(t, found)
	assert.Equal(t, calledVal, parentVal)
}
