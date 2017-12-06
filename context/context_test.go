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
	c = c.WithRequestID(expectedRequestID)
	c = c.WithActor(expectedUser)

	return c
}

func TestNew(t *testing.T) {
	actor := &models.FakeUser{}
	id := "555"

	c := New(context.Background())
	c = c.WithRequestID(id)
	c = c.WithActor(actor)

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

func TestWithActor(t *testing.T) {
	c := New(context.Background())
	new := c.WithActor(expectedUser)

	assert.Equal(t, expectedUser, c.Actor())
	assert.Equal(t, expectedUser, new.Actor())
	assert.Equal(t, c, new)
}

func TestWithRequestID(t *testing.T) {
	c := New(context.Background())
	new := c.WithRequestID(expectedRequestID)

	assert.Equal(t, expectedRequestID, c.RequestID())
	assert.Equal(t, expectedRequestID, new.RequestID())
	assert.Equal(t, c, new)
}

func TestWithValue(t *testing.T) {
	c1 := New(context.Background())
	new1 := c1.WithValue(ActorKey, expectedUser)

	assert.Equal(t, expectedUser, c1.Actor())
	assert.Equal(t, expectedUser, new1.Actor())
	assert.Equal(t, c1, new1)

	c2 := New(context.Background())
	new2 := c2.WithValue(IDKey, expectedRequestID)

	assert.Equal(t, expectedRequestID, c2.RequestID())
	assert.Equal(t, expectedRequestID, new2.RequestID())
	assert.Equal(t, c2, new2)

	c3 := New(context.Background())
	new3 := c3.WithValue("mnky", "fnky")

	assert.Equal(t, "fnky", c3.Value("mnky"))
	assert.Equal(t, "fnky", new3.Value("mnky"))
	assert.Equal(t, c3, new3)
}

func TestWithDeadline(t *testing.T) {
	c := New(context.Background())
	new, cancel := c.WithDeadline(time.Time{})
	assert.NotNil(t, new)
	assert.NotNil(t, cancel)
	defer cancel()

	assert.Equal(t, c, new)
}

func TestWithTimeout(t *testing.T) {
	c := New(context.Background())
	new, cancel := c.WithTimeout(time.Second * 5)
	assert.NotNil(t, new)
	assert.NotNil(t, cancel)
	defer cancel()

	assert.Equal(t, c, new)
}

func TestWithCancel(t *testing.T) {
	c, cancel := WithCancel(context.WithValue(context.Background(), "mnky", "fnky"))
	defer cancel()

	assert.NotNil(t, c)
	assert.NotNil(t, cancel)
	assert.Equal(t, "fnky", c.Value("mnky"))
}

func TestWithDeadlineConstructor(t *testing.T) {
	c, cancel := WithDeadline(context.WithValue(context.Background(), "mnky", "fnky"), time.Time{})
	defer cancel()

	assert.NotNil(t, c)
	assert.NotNil(t, cancel)
	assert.Equal(t, "fnky", c.Value("mnky"))
}

func TestWithTimeoutConstructor(t *testing.T) {
	c, cancel := WithTimeout(context.WithValue(context.Background(), "mnky", "fnky"), time.Second*5)
	defer cancel()

	assert.NotNil(t, c)
	assert.NotNil(t, cancel)
	assert.Equal(t, "fnky", c.Value("mnky"))
}

func TestWithActorConstructor(t *testing.T) {
	c := WithActor(context.Background(), expectedUser)
	assert.Equal(t, expectedUser, c.Actor())
}

func TestWithRequestIDConstructor(t *testing.T) {
	c := WithRequestID(context.Background(), expectedRequestID)
	assert.Equal(t, expectedRequestID, c.RequestID())
}

func TestWithValueConstructor(t *testing.T) {
	new1 := WithValue(context.Background(), ActorKey, expectedUser)
	assert.Equal(t, expectedUser, new1.Actor())

	new2 := WithValue(context.Background(), IDKey, expectedRequestID)
	assert.Equal(t, expectedRequestID, new2.RequestID())

	new3 := WithValue(context.Background(), "mnky", "fnky")
	assert.Equal(t, "fnky", new3.Value("mnky"))
}
