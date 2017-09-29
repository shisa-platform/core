package context

import (
	"context"
	"testing"
	"time"

	"github.com/percolate/shisa/log"
	"github.com/percolate/shisa/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type FakeUser struct {
	models.User
}

type FakeLogger struct {
	logx.Logger
	mock.Mock
}

type FakeContext struct {
	context.Context
	mock.Mock
	key, val interface{}
}

func (u *FakeUser) ID() string {
	return "123456"
}

func (u *FakeUser) String() string {
	return u.ID()
}

func (l *FakeLogger) Info(requestID, message string) {
	l.Called(requestID, message)
}

func (l *FakeLogger) Infof(requestID, format string, args ...interface{}) {
	l.Called(requestID, format, args[0])
}

func (l *FakeLogger) Error(requestID, message string) {
	l.Called(requestID, message)
}

func (l *FakeLogger) Errorf(requestID, format string, args ...interface{}) {
	l.Called(requestID, format, args[0])
}

func (l *FakeLogger) Trace(requestID, message string) {
	l.Called(requestID, message)
}

func (l *FakeLogger) Tracef(requestID, format string, args ...interface{}) {
	l.Called(requestID, format, args[0])
}

func (c *FakeContext) Deadline() (deadline time.Time, ok bool) {
	c.Called()
	return
}

func (c *FakeContext) Done() <-chan struct{} {
	c.Called()
	d := make(chan struct{})
	go func() {
		d <- struct{}{}
		close(d)
	}()
	return d
}

func (c *FakeContext) Err() error {
	c.Called()
	return nil
}

func (c *FakeContext) Value(key interface{}) interface{} {
	c.Called(key)
	return c.val
}

func createParentContext() *FakeContext {
	return &FakeContext{
		Context: context.Background(),
	}
}

func getContextForLogging(requestID string, logger *FakeLogger) *Context {
	actor := &FakeUser{}
	parent := createParentContext()

	c := New(parent, requestID, actor, logger)

	return c
}

func getContextForParent(parent *FakeContext) *Context {
	actor := &FakeUser{}
	logger := &FakeLogger{}

	c := New(parent, "999999", actor, logger)

	return c
}

func TestNew(t *testing.T) {
	actor := &FakeUser{}
	logger := &FakeLogger{}
	parent := createParentContext()
	id := "555"

	c := New(parent, id, actor, logger)

	assert.Equal(t, parent, c.Context)
	assert.Equal(t, id, c.RequestID)
	assert.Equal(t, actor, c.Actor)
	assert.Equal(t, logger, c.Logger)
}

func TestInfo(t *testing.T) {
	logtype := "Info"
	requestID := "555"
	message := "Hello test"
	logger := &FakeLogger{}
	logger.On(logtype, requestID, message).Return()
	c := getContextForLogging(requestID, logger)

	c.Info(message)

	logger.AssertCalled(t, logtype, requestID, message)
}

func TestInfof(t *testing.T) {
	logtype := "Infof"
	requestID := "555"
	format := "Hello test %s"
	name := "name"
	logger := &FakeLogger{}
	logger.On(logtype, requestID, format, name).Return()
	c := getContextForLogging(requestID, logger)

	c.Infof(format, name)

	logger.AssertCalled(t, logtype, requestID, format, name)
}

func TestError(t *testing.T) {
	logtype := "Error"
	requestID := "555"
	message := "Hello test"
	logger := &FakeLogger{}
	logger.On(logtype, requestID, message).Return()
	c := getContextForLogging(requestID, logger)

	c.Error(message)

	logger.AssertCalled(t, logtype, requestID, message)
}

func TestErrorf(t *testing.T) {
	logtype := "Errorf"
	requestID := "555"
	format := "Hello test %s"
	name := "name"
	logger := &FakeLogger{}
	logger.On(logtype, requestID, format, name).Return()
	c := getContextForLogging(requestID, logger)

	c.Errorf(format, name)

	logger.AssertCalled(t, logtype, requestID, format, name)
}

func TestTrace(t *testing.T) {
	logtype := "Trace"
	requestID := "555"
	message := "Hello test"
	logger := &FakeLogger{}
	logger.On(logtype, requestID, message).Return()
	c := getContextForLogging(requestID, logger)

	c.Trace(message)

	logger.AssertCalled(t, logtype, requestID, message)
}

func TestTracef(t *testing.T) {
	logtype := "Tracef"
	requestID := "555"
	format := "Hello test %s"
	name := "name"
	logger := &FakeLogger{}
	logger.On(logtype, requestID, format, name).Return()
	c := getContextForLogging(requestID, logger)

	c.Tracef(format, name)

	logger.AssertCalled(t, logtype, requestID, format, name)
}

func TestDeadline(t *testing.T) {
	parent := &FakeContext{}
	c := getContextForParent(parent)
	defaulttime := time.Time{}
	defaultok := false
	parent.On("Deadline").Return(defaulttime, defaultok)

	d, y := c.Deadline()

	parent.AssertCalled(t, "Deadline")
	assert.Equal(t, d, defaulttime)
	assert.Equal(t, y, defaultok)
}

func TestDone(t *testing.T) {
	parent := &FakeContext{}
	c := getContextForParent(parent)
	parent.On("Done").Return(struct{}{})

	d := <-c.Done()

	parent.AssertCalled(t, "Done")
	assert.Equal(t, d, struct{}{})
}

func TestErr(t *testing.T) {
	parent := &FakeContext{}
	c := getContextForParent(parent)
	parent.On("Err").Return(nil)

	err := c.Err()

	parent.AssertCalled(t, "Err")
	assert.Equal(t, err, nil)
}

func TestValueID(t *testing.T) {
	pkey := "ParentKey"
	pval := true
	parent := &FakeContext{
		key: pkey,
		val: pval,
	}
	parent.On("Value", pkey).Return(pval)
	c := getContextForParent(parent)

	idVal := c.Value(IDKey)
	actorVal := c.Value(ActorKey)
	loggerVal := c.Value(LoggerKey)
	parentVal := c.Value(pkey)

	assert.Equal(t, idVal, c.RequestID)
	assert.Equal(t, actorVal, c.Actor)
	assert.Equal(t, loggerVal, c.Logger)
	assert.Equal(t, parentVal, pval)
}
