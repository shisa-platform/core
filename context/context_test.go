package context

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/shisa-platform/core/models"

	"github.com/stretchr/testify/assert"
)

var _ context.Context = &FakeContext{}
var _ Context = &FakeContext{}

var (
	expectedUser      = &models.FakeUser{IDHook: func() string { return "123" }}
	expectedRequestID = "999999"
)

func getContextForParent(parent context.Context) Context {
	ctx := New(parent)
	ctx = ctx.WithRequestID(expectedRequestID)
	ctx = ctx.WithActor(expectedUser)

	return ctx
}

func TestNew(t *testing.T) {
	actor := &models.FakeUser{}
	id := "555"

	cut := New(context.Background())
	cut = cut.WithRequestID(id)
	cut = cut.WithActor(actor)

	assert.Equal(t, id, cut.RequestID())
	assert.Equal(t, actor, cut.Actor())
}

func TestDeadline(t *testing.T) {
	deadline := time.Time{}
	ok := false

	parent := &FakeContext{
		DeadlineHook: func() (deadline time.Time, ok bool) {
			return deadline, ok
		},
	}
	cut := getContextForParent(parent)

	d, y := cut.Deadline()

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
	cut := getContextForParent(parent)

	result := <-cut.Done()

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
	cut := getContextForParent(parent)

	result := cut.Err()

	parent.AssertErrCalledOnce(t)
	assert.Equal(t, result, err)
}

func TestValue(t *testing.T) {
	pkey := "ParentKey"
	pval := true

	parent := &FakeContext{
		ValueHook: func(key interface{}) (value interface{}) {
			if key == pkey {
				return pval
			}
			return
		},
	}

	cut := getContextForParent(parent)

	idVal := cut.Value(IDKey)
	actorVal := cut.Value(ActorKey)
	parentVal := cut.Value(pkey)

	assert.Equal(t, idVal, cut.RequestID())
	assert.Equal(t, actorVal, cut.Actor())

	parent.AssertValueCalledOnceWith(t, pkey)

	calledVal, found := parent.ValueResultsForCall(pkey)
	assert.True(t, found)
	assert.Equal(t, calledVal, parentVal)
}

func TestInheritedValue(t *testing.T) {
	pkey := "ParentKey"
	pval := true

	parent := &FakeContext{
		ValueHook: func(key interface{}) (value interface{}) {
			if key == pkey {
				return pval
			}
			return
		},
	}

	cut := getContextForParent(parent)

	cut = New(cut)

	assert.Equal(t, expectedUser, cut.Actor())
	assert.Equal(t, expectedUser, cut.Value(ActorKey).(models.User))
	assert.Equal(t, expectedRequestID, cut.RequestID())
	assert.Equal(t, expectedRequestID, cut.Value(IDKey).(string))
}

func TestWithParent(t *testing.T) {
	parent := context.WithValue(context.Background(), "foo", "bar")
	cut := New(context.Background())

	cut1 := cut.WithParent(parent)

	assert.Equal(t, "bar", cut1.Value("foo"))
}

func TestWithActor(t *testing.T) {
	cut := New(context.Background())
	new := cut.WithActor(expectedUser)

	assert.Equal(t, expectedUser, cut.Actor())
	assert.Equal(t, expectedUser, new.Actor())
	assert.Equal(t, cut, new)
}

func TestWithRequestID(t *testing.T) {
	cut := New(context.Background())
	new := cut.WithRequestID(expectedRequestID)

	assert.Equal(t, expectedRequestID, cut.RequestID())
	assert.Equal(t, expectedRequestID, new.RequestID())
	assert.Equal(t, cut, new)
}

func TestWithSpan(t *testing.T) {
	cut := New(context.Background())
	assert.Nil(t, cut.Span())

	tracer := opentracing.NoopTracer{}
	parent := tracer.StartSpan("test")
	defer parent.Finish()

	ctx := cut.WithSpan(parent)

	assert.Equal(t, parent, ctx.Span())
}

func TestInheritedSpan(t *testing.T) {
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "test")
	defer span.Finish()

	cut := New(ctx)

	assert.Equal(t, span, cut.Span())

	cut = WithSpan(context.Background(), span)
	cut = New(cut)

	assert.Equal(t, span, cut.Span())
}

func TestStartSpan(t *testing.T) {
	cut := New(context.Background())
	span := cut.StartSpan("test")
	assert.NotNil(t, span)
	defer span.Finish()
	assert.Nil(t, cut.Span())

	cut.WithSpan(span)
	assert.True(t, span == cut.Span())

	span2 := cut.StartSpan("test2")
	defer span2.Finish()

	assert.NotNil(t, span2)
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

	span := opentracing.StartSpan("test")
	c3 := New(context.Background())
	new3 := c3.WithValue(SpanKey, span)

	assert.Equal(t, span, c3.Span())
	assert.Equal(t, span, new3.Span())
	assert.Equal(t, c3, new3)

	c4 := New(context.Background())
	new4 := c4.WithValue("mnky", "fnky")

	assert.Equal(t, "fnky", c4.Value("mnky"))
	assert.Equal(t, "fnky", new4.Value("mnky"))
	assert.Equal(t, c4, new4)
}

func TestWithCancel(t *testing.T) {
	cut := New(context.Background())
	new, cancel := cut.WithCancel()
	assert.NotNil(t, new)
	assert.NotNil(t, cancel)
	defer cancel()

	assert.Equal(t, cut, new)
}

func TestWithDeadline(t *testing.T) {
	cut := New(context.Background())
	new, cancel := cut.WithDeadline(time.Time{})
	assert.NotNil(t, new)
	assert.NotNil(t, cancel)
	defer cancel()

	assert.Equal(t, cut, new)
}

func TestWithTimeout(t *testing.T) {
	cut := New(context.Background())
	new, cancel := cut.WithTimeout(time.Second * 5)
	assert.NotNil(t, new)
	assert.NotNil(t, cancel)
	defer cancel()

	assert.Equal(t, cut, new)
}

func TestWithCancelConstructor(t *testing.T) {
	cut, cancel := WithCancel(context.WithValue(context.Background(), "mnky", "fnky"))
	defer cancel()

	assert.NotNil(t, cut)
	assert.NotNil(t, cancel)
	assert.Equal(t, "fnky", cut.Value("mnky"))
}

func TestWithDeadlineConstructor(t *testing.T) {
	cut, cancel := WithDeadline(context.WithValue(context.Background(), "mnky", "fnky"), time.Time{})
	defer cancel()

	assert.NotNil(t, cut)
	assert.NotNil(t, cancel)
	assert.Equal(t, "fnky", cut.Value("mnky"))
}

func TestWithTimeoutConstructor(t *testing.T) {
	cut, cancel := WithTimeout(context.WithValue(context.Background(), "mnky", "fnky"), time.Second*5)
	defer cancel()

	assert.NotNil(t, cut)
	assert.NotNil(t, cancel)
	assert.Equal(t, "fnky", cut.Value("mnky"))
}

func TestWithActorConstructor(t *testing.T) {
	cut := WithActor(context.Background(), expectedUser)
	assert.Equal(t, expectedUser, cut.Actor())
}

func TestWithRequestIDConstructor(t *testing.T) {
	cut := WithRequestID(context.Background(), expectedRequestID)
	assert.Equal(t, expectedRequestID, cut.RequestID())
}

func TestWithSpanConstructor(t *testing.T) {
	span := opentracing.StartSpan("test")
	cut := WithSpan(context.Background(), span)

	assert.Equal(t, span, cut.Span())
}

func TestStartSpanConstructor(t *testing.T) {
	parent := New(context.Background())
	span, cut := StartSpan(parent, "test")

	assert.Equal(t, span, cut.Span())
}

func TestWithValueConstructor(t *testing.T) {
	new1 := WithValue(context.Background(), ActorKey, expectedUser)
	assert.Equal(t, expectedUser, new1.Actor())

	new2 := WithValue(context.Background(), IDKey, expectedRequestID)
	assert.Equal(t, expectedRequestID, new2.RequestID())

	span := opentracing.StartSpan("test")
	new3 := WithValue(context.Background(), SpanKey, span)
	assert.Equal(t, span, new3.Span())

	new4 := WithValue(context.Background(), "mnky", "fnky")
	assert.Equal(t, "fnky", new4.Value("mnky"))
}
