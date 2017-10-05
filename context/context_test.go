package context

import (
	"context"
	"testing"
	"time"

	"errors"
	"github.com/percolate/shisa/context/contexttest"
	"github.com/percolate/shisa/log/logtest"
	"github.com/percolate/shisa/models/modelstest"
	"github.com/stretchr/testify/assert"
)

var _ context.Context = &contexttest.FakeContext{}

func getContextForLogging(requestID string, logger *logtest.FakeLogger) *Context {
	actor := &modelstest.FakeUser{}
	parent := &contexttest.FakeContext{}

	c := New(parent, requestID, actor, logger)

	return c
}

func getContextForParent(parent context.Context) *Context {
	actor := &modelstest.FakeUser{}
	logger := &logtest.FakeLogger{}

	c := New(parent, "999999", actor, logger)

	return c
}

func TestNew(t *testing.T) {
	actor := &modelstest.FakeUser{}
	logger := &logtest.FakeLogger{}
	parent := &contexttest.FakeContext{}
	id := "555"

	c := New(parent, id, actor, logger)

	assert.Equal(t, parent, c.Context)
	assert.Equal(t, id, c.RequestID)
	assert.Equal(t, actor, c.Actor)
	assert.Equal(t, logger, c.Logger)
}

func TestInfo(t *testing.T) {
	requestID := "555"
	message := "Hello test"

	logger := &logtest.FakeLogger{
		InfoHook: func(requestID string, message string) {},
	}

	c := getContextForLogging(requestID, logger)

	c.Info(message)

	assert.Equal(t, len(logger.InfoCalls), 1)

	invocation := logger.InfoCalls[0]

	assert.Equal(t, invocation.Parameters.RequestID, requestID)
	assert.Equal(t, invocation.Parameters.Message, message)
}

func TestInfof(t *testing.T) {
	requestID := "555"
	format := "Hello test %s"

	// lol no generics
	s := []string{"arg"}
	args := make([]interface{}, len(s))
	for i, v := range s {
		args[i] = v
	}

	logger := &logtest.FakeLogger{
		InfofHook: func(requestID string, format string, args ...interface{}) {},
	}

	c := getContextForLogging(requestID, logger)

	c.Infof(format, args...)

	assert.Equal(t, len(logger.InfofCalls), 1)

	invocation := logger.InfofCalls[0]

	assert.Equal(t, invocation.Parameters.RequestID, requestID)
	assert.Equal(t, invocation.Parameters.Format, format)
	assert.Equal(t, invocation.Parameters.Args, args)
}

func TestError(t *testing.T) {
	requestID := "555"
	message := "Hello test"

	logger := &logtest.FakeLogger{
		ErrorHook: func(requestID string, message string) {},
	}

	c := getContextForLogging(requestID, logger)

	c.Error(message)

	assert.Equal(t, len(logger.ErrorCalls), 1)

	invocation := logger.ErrorCalls[0]

	assert.Equal(t, invocation.Parameters.RequestID, requestID)
	assert.Equal(t, invocation.Parameters.Message, message)
}

func TestErrorf(t *testing.T) {
	requestID := "555"
	format := "Hello test %s"

	// lol no generics
	s := []string{"arg"}
	args := make([]interface{}, len(s))
	for i, v := range s {
		args[i] = v
	}

	logger := &logtest.FakeLogger{
		ErrorfHook: func(requestID string, format string, args ...interface{}) {},
	}

	c := getContextForLogging(requestID, logger)

	c.Errorf(format, args...)

	assert.Equal(t, len(logger.ErrorfCalls), 1)

	invocation := logger.ErrorfCalls[0]

	assert.Equal(t, invocation.Parameters.RequestID, requestID)
	assert.Equal(t, invocation.Parameters.Format, format)
	assert.Equal(t, invocation.Parameters.Args, args)
}

func TestTrace(t *testing.T) {
	requestID := "555"
	message := "Hello test"

	logger := &logtest.FakeLogger{
		TraceHook: func(requestID string, message string) {},
	}

	c := getContextForLogging(requestID, logger)

	c.Trace(message)

	assert.Equal(t, len(logger.TraceCalls), 1)

	invocation := logger.TraceCalls[0]

	assert.Equal(t, invocation.Parameters.RequestID, requestID)
	assert.Equal(t, invocation.Parameters.Message, message)
}

func TestTracef(t *testing.T) {
	requestID := "555"
	format := "Hello test %s"

	// lol no generics
	s := []string{"arg"}
	args := make([]interface{}, len(s))
	for i, v := range s {
		args[i] = v
	}

	logger := &logtest.FakeLogger{
		TracefHook: func(requestID string, format string, args ...interface{}) {},
	}

	c := getContextForLogging(requestID, logger)

	c.Tracef(format, args...)

	assert.Equal(t, len(logger.TracefCalls), 1)

	invocation := logger.TracefCalls[0]

	assert.Equal(t, invocation.Parameters.RequestID, requestID)
	assert.Equal(t, invocation.Parameters.Format, format)
	assert.Equal(t, invocation.Parameters.Args, args)
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

	assert.Equal(t, len(parent.DeadlineCalls), 1)

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

	assert.Equal(t, len(parent.DoneCalls), 1)
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

	assert.Equal(t, len(parent.ErrCalls), 1)
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
	loggerVal := c.Value(LoggerKey)
	parentVal := c.Value(pkey)

	assert.Equal(t, idVal, c.RequestID)
	assert.Equal(t, actorVal, c.Actor)
	assert.Equal(t, loggerVal, c.Logger)

	assert.Equal(t, len(parent.ValueCalls), 1)
	invocation := parent.ValueCalls[0]
	assert.Equal(t, invocation.Parameters.Key, pkey)
	assert.Equal(t, invocation.Results.Ret0, parentVal)
}
