package context

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/percolate/shisa/models"

	"github.com/stretchr/testify/assert"
)

var _ context.Context = &FakeContext{}
var _ Context = &FakeContext{}

var (
	expectedUser      = &models.FakeUser{IDHook: func() string { return "123" }}
	expectedRequestID = "999999"
)

func getContextForParent(parent context.Context) Context {
	c := New(parent)
	c.SetRequestID(expectedRequestID)
	c.SetActor(expectedUser)
	return c
}

func TestNew(t *testing.T) {
	actor := &models.FakeUser{}
	id := "555"

	c := New(context.Background())
	c.SetRequestID(id)
	c.SetActor(actor)

	assert.Equal(t, id, c.RequestID())
	assert.Equal(t, actor, c.Actor())
}

func TestDeadline(t *testing.T) {
	deadline := time.Time{}
	ok := false

	parent := &FakeContext{
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

	parent := &FakeContext{
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

	parent := &FakeContext{
		ErrHook: func() (ret0 error) {
			return err
		},
	}
	c := getContextForParent(parent)

	result := c.Err()

	parent.AssertErrCalledOnce(t)
	assert.Equal(t, result, err)
}

func TestValue(t *testing.T) {
	pkey := "ParentKey"
	pval := true

	parent := &FakeContext{
		ValueHook: func(key interface{}) (ret0 interface{}) {
			return pval
		},
	}

	c := getContextForParent(parent)

	idVal := c.Value(IDKey)
	actorVal := c.Value(ActorKey)
	parentVal := c.Value(pkey)

	assert.Equal(t, idVal, c.RequestID())
	assert.Equal(t, actorVal, c.Actor())

	parent.AssertValueCalledOnceWith(t, pkey)

	calledVal, found := parent.ValueResultsForCall(pkey)
	assert.True(t, found)
	assert.Equal(t, calledVal, parentVal)
}
