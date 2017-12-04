// generated by "charlatan -output=./context_charlatan.go Context".  DO NOT EDIT.

package context

import (
	"reflect"
	"testing"
	"time"

	"github.com/percolate/shisa/models"
)

// DeadlineInvocation represents a single call of FakeContext.Deadline
type DeadlineInvocation struct {
	Results struct {
		Deadline time.Time
		Ok       bool
	}
}

// DoneInvocation represents a single call of FakeContext.Done
type DoneInvocation struct {
	Results struct {
		Ident1 <-chan struct{}
	}
}

// ErrInvocation represents a single call of FakeContext.Err
type ErrInvocation struct {
	Results struct {
		Ident2 error
	}
}

// ValueInvocation represents a single call of FakeContext.Value
type ValueInvocation struct {
	Parameters struct {
		Key interface{}
	}
	Results struct {
		Ident3 interface{}
	}
}

// RequestIDInvocation represents a single call of FakeContext.RequestID
type RequestIDInvocation struct {
	Results struct {
		Ident6 string
	}
}

// SetRequestIDInvocation represents a single call of FakeContext.SetRequestID
type SetRequestIDInvocation struct {
	Parameters struct {
		V string
	}
}

// ActorInvocation represents a single call of FakeContext.Actor
type ActorInvocation struct {
	Results struct {
		Ident7 models.User
	}
}

// SetActorInvocation represents a single call of FakeContext.SetActor
type SetActorInvocation struct {
	Parameters struct {
		V models.User
	}
}

/*
FakeContext is a mock implementation of Context for testing.
Use it in your tests as in this example:

	package example

	func TestWithContext(t *testing.T) {
		f := &context.FakeContext{
			DeadlineHook: func() (deadline time.Time, ok bool) {
				// ensure parameters meet expections, signal errors using t, etc
				return
			},
		}

		// test code goes here ...

		// assert state of FakeDeadline ...
		f.AssertDeadlineCalledOnce(t)
	}

Create anonymous function implementations for only those interface methods that
should be called in the code under test.  This will force a panic if any
unexpected calls are made to FakeDeadline.
*/
type FakeContext struct {
	DeadlineHook     func() (time.Time, bool)
	DoneHook         func() <-chan struct{}
	ErrHook          func() error
	ValueHook        func(interface{}) interface{}
	RequestIDHook    func() string
	SetRequestIDHook func(string)
	ActorHook        func() models.User
	SetActorHook     func(models.User)

	DeadlineCalls     []*DeadlineInvocation
	DoneCalls         []*DoneInvocation
	ErrCalls          []*ErrInvocation
	ValueCalls        []*ValueInvocation
	RequestIDCalls    []*RequestIDInvocation
	SetRequestIDCalls []*SetRequestIDInvocation
	ActorCalls        []*ActorInvocation
	SetActorCalls     []*SetActorInvocation
}

// NewFakeContextDefaultPanic returns an instance of FakeContext with all hooks configured to panic
func NewFakeContextDefaultPanic() *FakeContext {
	return &FakeContext{
		DeadlineHook: func() (deadline time.Time, ok bool) {
			panic("Unexpected call to Context.Deadline")
		},
		DoneHook: func() (ident1 <-chan struct{}) {
			panic("Unexpected call to Context.Done")
		},
		ErrHook: func() (ident2 error) {
			panic("Unexpected call to Context.Err")
		},
		ValueHook: func(interface{}) (ident3 interface{}) {
			panic("Unexpected call to Context.Value")
		},
		RequestIDHook: func() (ident6 string) {
			panic("Unexpected call to Context.RequestID")
		},
		SetRequestIDHook: func(string) {
			panic("Unexpected call to Context.SetRequestID")
		},
		ActorHook: func() (ident7 models.User) {
			panic("Unexpected call to Context.Actor")
		},
		SetActorHook: func(models.User) {
			panic("Unexpected call to Context.SetActor")
		},
	}
}

// NewFakeContextDefaultFatal returns an instance of FakeContext with all hooks configured to call t.Fatal
func NewFakeContextDefaultFatal(t *testing.T) *FakeContext {
	return &FakeContext{
		DeadlineHook: func() (deadline time.Time, ok bool) {
			t.Fatal("Unexpected call to Context.Deadline")
			return
		},
		DoneHook: func() (ident1 <-chan struct{}) {
			t.Fatal("Unexpected call to Context.Done")
			return
		},
		ErrHook: func() (ident2 error) {
			t.Fatal("Unexpected call to Context.Err")
			return
		},
		ValueHook: func(interface{}) (ident3 interface{}) {
			t.Fatal("Unexpected call to Context.Value")
			return
		},
		RequestIDHook: func() (ident6 string) {
			t.Fatal("Unexpected call to Context.RequestID")
			return
		},
		SetRequestIDHook: func(string) {
			t.Fatal("Unexpected call to Context.SetRequestID")
			return
		},
		ActorHook: func() (ident7 models.User) {
			t.Fatal("Unexpected call to Context.Actor")
			return
		},
		SetActorHook: func(models.User) {
			t.Fatal("Unexpected call to Context.SetActor")
			return
		},
	}
}

// NewFakeContextDefaultError returns an instance of FakeContext with all hooks configured to call t.Error
func NewFakeContextDefaultError(t *testing.T) *FakeContext {
	return &FakeContext{
		DeadlineHook: func() (deadline time.Time, ok bool) {
			t.Error("Unexpected call to Context.Deadline")
			return
		},
		DoneHook: func() (ident1 <-chan struct{}) {
			t.Error("Unexpected call to Context.Done")
			return
		},
		ErrHook: func() (ident2 error) {
			t.Error("Unexpected call to Context.Err")
			return
		},
		ValueHook: func(interface{}) (ident3 interface{}) {
			t.Error("Unexpected call to Context.Value")
			return
		},
		RequestIDHook: func() (ident6 string) {
			t.Error("Unexpected call to Context.RequestID")
			return
		},
		SetRequestIDHook: func(string) {
			t.Error("Unexpected call to Context.SetRequestID")
			return
		},
		ActorHook: func() (ident7 models.User) {
			t.Error("Unexpected call to Context.Actor")
			return
		},
		SetActorHook: func(models.User) {
			t.Error("Unexpected call to Context.SetActor")
			return
		},
	}
}

func (_f1 *FakeContext) Deadline() (deadline time.Time, ok bool) {
	invocation := new(DeadlineInvocation)

	deadline, ok = _f1.DeadlineHook()

	invocation.Results.Deadline = deadline
	invocation.Results.Ok = ok

	_f1.DeadlineCalls = append(_f1.DeadlineCalls, invocation)

	return
}

// DeadlineCalled returns true if FakeContext.Deadline was called
func (f *FakeContext) DeadlineCalled() bool {
	return len(f.DeadlineCalls) != 0
}

// AssertDeadlineCalled calls t.Error if FakeContext.Deadline was not called
func (f *FakeContext) AssertDeadlineCalled(t *testing.T) {
	t.Helper()
	if len(f.DeadlineCalls) == 0 {
		t.Error("FakeContext.Deadline not called, expected at least one")
	}
}

// DeadlineNotCalled returns true if FakeContext.Deadline was not called
func (f *FakeContext) DeadlineNotCalled() bool {
	return len(f.DeadlineCalls) == 0
}

// AssertDeadlineNotCalled calls t.Error if FakeContext.Deadline was called
func (f *FakeContext) AssertDeadlineNotCalled(t *testing.T) {
	t.Helper()
	if len(f.DeadlineCalls) != 0 {
		t.Error("FakeContext.Deadline called, expected none")
	}
}

// DeadlineCalledOnce returns true if FakeContext.Deadline was called exactly once
func (f *FakeContext) DeadlineCalledOnce() bool {
	return len(f.DeadlineCalls) == 1
}

// AssertDeadlineCalledOnce calls t.Error if FakeContext.Deadline was not called exactly once
func (f *FakeContext) AssertDeadlineCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.DeadlineCalls) != 1 {
		t.Errorf("FakeContext.Deadline called %d times, expected 1", len(f.DeadlineCalls))
	}
}

// DeadlineCalledN returns true if FakeContext.Deadline was called at least n times
func (f *FakeContext) DeadlineCalledN(n int) bool {
	return len(f.DeadlineCalls) >= n
}

// AssertDeadlineCalledN calls t.Error if FakeContext.Deadline was called less than n times
func (f *FakeContext) AssertDeadlineCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.DeadlineCalls) < n {
		t.Errorf("FakeContext.Deadline called %d times, expected >= %d", len(f.DeadlineCalls), n)
	}
}

func (_f2 *FakeContext) Done() (ident1 <-chan struct{}) {
	invocation := new(DoneInvocation)

	ident1 = _f2.DoneHook()

	invocation.Results.Ident1 = ident1

	_f2.DoneCalls = append(_f2.DoneCalls, invocation)

	return
}

// DoneCalled returns true if FakeContext.Done was called
func (f *FakeContext) DoneCalled() bool {
	return len(f.DoneCalls) != 0
}

// AssertDoneCalled calls t.Error if FakeContext.Done was not called
func (f *FakeContext) AssertDoneCalled(t *testing.T) {
	t.Helper()
	if len(f.DoneCalls) == 0 {
		t.Error("FakeContext.Done not called, expected at least one")
	}
}

// DoneNotCalled returns true if FakeContext.Done was not called
func (f *FakeContext) DoneNotCalled() bool {
	return len(f.DoneCalls) == 0
}

// AssertDoneNotCalled calls t.Error if FakeContext.Done was called
func (f *FakeContext) AssertDoneNotCalled(t *testing.T) {
	t.Helper()
	if len(f.DoneCalls) != 0 {
		t.Error("FakeContext.Done called, expected none")
	}
}

// DoneCalledOnce returns true if FakeContext.Done was called exactly once
func (f *FakeContext) DoneCalledOnce() bool {
	return len(f.DoneCalls) == 1
}

// AssertDoneCalledOnce calls t.Error if FakeContext.Done was not called exactly once
func (f *FakeContext) AssertDoneCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.DoneCalls) != 1 {
		t.Errorf("FakeContext.Done called %d times, expected 1", len(f.DoneCalls))
	}
}

// DoneCalledN returns true if FakeContext.Done was called at least n times
func (f *FakeContext) DoneCalledN(n int) bool {
	return len(f.DoneCalls) >= n
}

// AssertDoneCalledN calls t.Error if FakeContext.Done was called less than n times
func (f *FakeContext) AssertDoneCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.DoneCalls) < n {
		t.Errorf("FakeContext.Done called %d times, expected >= %d", len(f.DoneCalls), n)
	}
}

func (_f3 *FakeContext) Err() (ident2 error) {
	invocation := new(ErrInvocation)

	ident2 = _f3.ErrHook()

	invocation.Results.Ident2 = ident2

	_f3.ErrCalls = append(_f3.ErrCalls, invocation)

	return
}

// ErrCalled returns true if FakeContext.Err was called
func (f *FakeContext) ErrCalled() bool {
	return len(f.ErrCalls) != 0
}

// AssertErrCalled calls t.Error if FakeContext.Err was not called
func (f *FakeContext) AssertErrCalled(t *testing.T) {
	t.Helper()
	if len(f.ErrCalls) == 0 {
		t.Error("FakeContext.Err not called, expected at least one")
	}
}

// ErrNotCalled returns true if FakeContext.Err was not called
func (f *FakeContext) ErrNotCalled() bool {
	return len(f.ErrCalls) == 0
}

// AssertErrNotCalled calls t.Error if FakeContext.Err was called
func (f *FakeContext) AssertErrNotCalled(t *testing.T) {
	t.Helper()
	if len(f.ErrCalls) != 0 {
		t.Error("FakeContext.Err called, expected none")
	}
}

// ErrCalledOnce returns true if FakeContext.Err was called exactly once
func (f *FakeContext) ErrCalledOnce() bool {
	return len(f.ErrCalls) == 1
}

// AssertErrCalledOnce calls t.Error if FakeContext.Err was not called exactly once
func (f *FakeContext) AssertErrCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.ErrCalls) != 1 {
		t.Errorf("FakeContext.Err called %d times, expected 1", len(f.ErrCalls))
	}
}

// ErrCalledN returns true if FakeContext.Err was called at least n times
func (f *FakeContext) ErrCalledN(n int) bool {
	return len(f.ErrCalls) >= n
}

// AssertErrCalledN calls t.Error if FakeContext.Err was called less than n times
func (f *FakeContext) AssertErrCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.ErrCalls) < n {
		t.Errorf("FakeContext.Err called %d times, expected >= %d", len(f.ErrCalls), n)
	}
}

func (_f4 *FakeContext) Value(key interface{}) (ident3 interface{}) {
	invocation := new(ValueInvocation)

	invocation.Parameters.Key = key

	ident3 = _f4.ValueHook(key)

	invocation.Results.Ident3 = ident3

	_f4.ValueCalls = append(_f4.ValueCalls, invocation)

	return
}

// ValueCalled returns true if FakeContext.Value was called
func (f *FakeContext) ValueCalled() bool {
	return len(f.ValueCalls) != 0
}

// AssertValueCalled calls t.Error if FakeContext.Value was not called
func (f *FakeContext) AssertValueCalled(t *testing.T) {
	t.Helper()
	if len(f.ValueCalls) == 0 {
		t.Error("FakeContext.Value not called, expected at least one")
	}
}

// ValueNotCalled returns true if FakeContext.Value was not called
func (f *FakeContext) ValueNotCalled() bool {
	return len(f.ValueCalls) == 0
}

// AssertValueNotCalled calls t.Error if FakeContext.Value was called
func (f *FakeContext) AssertValueNotCalled(t *testing.T) {
	t.Helper()
	if len(f.ValueCalls) != 0 {
		t.Error("FakeContext.Value called, expected none")
	}
}

// ValueCalledOnce returns true if FakeContext.Value was called exactly once
func (f *FakeContext) ValueCalledOnce() bool {
	return len(f.ValueCalls) == 1
}

// AssertValueCalledOnce calls t.Error if FakeContext.Value was not called exactly once
func (f *FakeContext) AssertValueCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.ValueCalls) != 1 {
		t.Errorf("FakeContext.Value called %d times, expected 1", len(f.ValueCalls))
	}
}

// ValueCalledN returns true if FakeContext.Value was called at least n times
func (f *FakeContext) ValueCalledN(n int) bool {
	return len(f.ValueCalls) >= n
}

// AssertValueCalledN calls t.Error if FakeContext.Value was called less than n times
func (f *FakeContext) AssertValueCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.ValueCalls) < n {
		t.Errorf("FakeContext.Value called %d times, expected >= %d", len(f.ValueCalls), n)
	}
}

// ValueCalledWith returns true if FakeContext.Value was called with the given values
func (_f5 *FakeContext) ValueCalledWith(key interface{}) (found bool) {
	for _, call := range _f5.ValueCalls {
		if reflect.DeepEqual(call.Parameters.Key, key) {
			found = true
			break
		}
	}

	return
}

// AssertValueCalledWith calls t.Error if FakeContext.Value was not called with the given values
func (_f6 *FakeContext) AssertValueCalledWith(t *testing.T, key interface{}) {
	t.Helper()
	var found bool
	for _, call := range _f6.ValueCalls {
		if reflect.DeepEqual(call.Parameters.Key, key) {
			found = true
			break
		}
	}

	if !found {
		t.Error("FakeContext.Value not called with expected parameters")
	}
}

// ValueCalledOnceWith returns true if FakeContext.Value was called exactly once with the given values
func (_f7 *FakeContext) ValueCalledOnceWith(key interface{}) bool {
	var count int
	for _, call := range _f7.ValueCalls {
		if reflect.DeepEqual(call.Parameters.Key, key) {
			count++
		}
	}

	return count == 1
}

// AssertValueCalledOnceWith calls t.Error if FakeContext.Value was not called exactly once with the given values
func (_f8 *FakeContext) AssertValueCalledOnceWith(t *testing.T, key interface{}) {
	t.Helper()
	var count int
	for _, call := range _f8.ValueCalls {
		if reflect.DeepEqual(call.Parameters.Key, key) {
			count++
		}
	}

	if count != 1 {
		t.Errorf("FakeContext.Value called %d times with expected parameters, expected one", count)
	}
}

// ValueResultsForCall returns the result values for the first call to FakeContext.Value with the given values
func (_f9 *FakeContext) ValueResultsForCall(key interface{}) (ident3 interface{}, found bool) {
	for _, call := range _f9.ValueCalls {
		if reflect.DeepEqual(call.Parameters.Key, key) {
			ident3 = call.Results.Ident3
			found = true
			break
		}
	}

	return
}

func (_f10 *FakeContext) RequestID() (ident6 string) {
	invocation := new(RequestIDInvocation)

	ident6 = _f10.RequestIDHook()

	invocation.Results.Ident6 = ident6

	_f10.RequestIDCalls = append(_f10.RequestIDCalls, invocation)

	return
}

// RequestIDCalled returns true if FakeContext.RequestID was called
func (f *FakeContext) RequestIDCalled() bool {
	return len(f.RequestIDCalls) != 0
}

// AssertRequestIDCalled calls t.Error if FakeContext.RequestID was not called
func (f *FakeContext) AssertRequestIDCalled(t *testing.T) {
	t.Helper()
	if len(f.RequestIDCalls) == 0 {
		t.Error("FakeContext.RequestID not called, expected at least one")
	}
}

// RequestIDNotCalled returns true if FakeContext.RequestID was not called
func (f *FakeContext) RequestIDNotCalled() bool {
	return len(f.RequestIDCalls) == 0
}

// AssertRequestIDNotCalled calls t.Error if FakeContext.RequestID was called
func (f *FakeContext) AssertRequestIDNotCalled(t *testing.T) {
	t.Helper()
	if len(f.RequestIDCalls) != 0 {
		t.Error("FakeContext.RequestID called, expected none")
	}
}

// RequestIDCalledOnce returns true if FakeContext.RequestID was called exactly once
func (f *FakeContext) RequestIDCalledOnce() bool {
	return len(f.RequestIDCalls) == 1
}

// AssertRequestIDCalledOnce calls t.Error if FakeContext.RequestID was not called exactly once
func (f *FakeContext) AssertRequestIDCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.RequestIDCalls) != 1 {
		t.Errorf("FakeContext.RequestID called %d times, expected 1", len(f.RequestIDCalls))
	}
}

// RequestIDCalledN returns true if FakeContext.RequestID was called at least n times
func (f *FakeContext) RequestIDCalledN(n int) bool {
	return len(f.RequestIDCalls) >= n
}

// AssertRequestIDCalledN calls t.Error if FakeContext.RequestID was called less than n times
func (f *FakeContext) AssertRequestIDCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.RequestIDCalls) < n {
		t.Errorf("FakeContext.RequestID called %d times, expected >= %d", len(f.RequestIDCalls), n)
	}
}

func (_f11 *FakeContext) SetRequestID(v string) {
	invocation := new(SetRequestIDInvocation)

	invocation.Parameters.V = v

	_f11.SetRequestIDHook(v)

	_f11.SetRequestIDCalls = append(_f11.SetRequestIDCalls, invocation)

	return
}

// SetRequestIDCalled returns true if FakeContext.SetRequestID was called
func (f *FakeContext) SetRequestIDCalled() bool {
	return len(f.SetRequestIDCalls) != 0
}

// AssertSetRequestIDCalled calls t.Error if FakeContext.SetRequestID was not called
func (f *FakeContext) AssertSetRequestIDCalled(t *testing.T) {
	t.Helper()
	if len(f.SetRequestIDCalls) == 0 {
		t.Error("FakeContext.SetRequestID not called, expected at least one")
	}
}

// SetRequestIDNotCalled returns true if FakeContext.SetRequestID was not called
func (f *FakeContext) SetRequestIDNotCalled() bool {
	return len(f.SetRequestIDCalls) == 0
}

// AssertSetRequestIDNotCalled calls t.Error if FakeContext.SetRequestID was called
func (f *FakeContext) AssertSetRequestIDNotCalled(t *testing.T) {
	t.Helper()
	if len(f.SetRequestIDCalls) != 0 {
		t.Error("FakeContext.SetRequestID called, expected none")
	}
}

// SetRequestIDCalledOnce returns true if FakeContext.SetRequestID was called exactly once
func (f *FakeContext) SetRequestIDCalledOnce() bool {
	return len(f.SetRequestIDCalls) == 1
}

// AssertSetRequestIDCalledOnce calls t.Error if FakeContext.SetRequestID was not called exactly once
func (f *FakeContext) AssertSetRequestIDCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.SetRequestIDCalls) != 1 {
		t.Errorf("FakeContext.SetRequestID called %d times, expected 1", len(f.SetRequestIDCalls))
	}
}

// SetRequestIDCalledN returns true if FakeContext.SetRequestID was called at least n times
func (f *FakeContext) SetRequestIDCalledN(n int) bool {
	return len(f.SetRequestIDCalls) >= n
}

// AssertSetRequestIDCalledN calls t.Error if FakeContext.SetRequestID was called less than n times
func (f *FakeContext) AssertSetRequestIDCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.SetRequestIDCalls) < n {
		t.Errorf("FakeContext.SetRequestID called %d times, expected >= %d", len(f.SetRequestIDCalls), n)
	}
}

// SetRequestIDCalledWith returns true if FakeContext.SetRequestID was called with the given values
func (_f12 *FakeContext) SetRequestIDCalledWith(v string) (found bool) {
	for _, call := range _f12.SetRequestIDCalls {
		if reflect.DeepEqual(call.Parameters.V, v) {
			found = true
			break
		}
	}

	return
}

// AssertSetRequestIDCalledWith calls t.Error if FakeContext.SetRequestID was not called with the given values
func (_f13 *FakeContext) AssertSetRequestIDCalledWith(t *testing.T, v string) {
	t.Helper()
	var found bool
	for _, call := range _f13.SetRequestIDCalls {
		if reflect.DeepEqual(call.Parameters.V, v) {
			found = true
			break
		}
	}

	if !found {
		t.Error("FakeContext.SetRequestID not called with expected parameters")
	}
}

// SetRequestIDCalledOnceWith returns true if FakeContext.SetRequestID was called exactly once with the given values
func (_f14 *FakeContext) SetRequestIDCalledOnceWith(v string) bool {
	var count int
	for _, call := range _f14.SetRequestIDCalls {
		if reflect.DeepEqual(call.Parameters.V, v) {
			count++
		}
	}

	return count == 1
}

// AssertSetRequestIDCalledOnceWith calls t.Error if FakeContext.SetRequestID was not called exactly once with the given values
func (_f15 *FakeContext) AssertSetRequestIDCalledOnceWith(t *testing.T, v string) {
	t.Helper()
	var count int
	for _, call := range _f15.SetRequestIDCalls {
		if reflect.DeepEqual(call.Parameters.V, v) {
			count++
		}
	}

	if count != 1 {
		t.Errorf("FakeContext.SetRequestID called %d times with expected parameters, expected one", count)
	}
}

func (_f16 *FakeContext) Actor() (ident7 models.User) {
	invocation := new(ActorInvocation)

	ident7 = _f16.ActorHook()

	invocation.Results.Ident7 = ident7

	_f16.ActorCalls = append(_f16.ActorCalls, invocation)

	return
}

// ActorCalled returns true if FakeContext.Actor was called
func (f *FakeContext) ActorCalled() bool {
	return len(f.ActorCalls) != 0
}

// AssertActorCalled calls t.Error if FakeContext.Actor was not called
func (f *FakeContext) AssertActorCalled(t *testing.T) {
	t.Helper()
	if len(f.ActorCalls) == 0 {
		t.Error("FakeContext.Actor not called, expected at least one")
	}
}

// ActorNotCalled returns true if FakeContext.Actor was not called
func (f *FakeContext) ActorNotCalled() bool {
	return len(f.ActorCalls) == 0
}

// AssertActorNotCalled calls t.Error if FakeContext.Actor was called
func (f *FakeContext) AssertActorNotCalled(t *testing.T) {
	t.Helper()
	if len(f.ActorCalls) != 0 {
		t.Error("FakeContext.Actor called, expected none")
	}
}

// ActorCalledOnce returns true if FakeContext.Actor was called exactly once
func (f *FakeContext) ActorCalledOnce() bool {
	return len(f.ActorCalls) == 1
}

// AssertActorCalledOnce calls t.Error if FakeContext.Actor was not called exactly once
func (f *FakeContext) AssertActorCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.ActorCalls) != 1 {
		t.Errorf("FakeContext.Actor called %d times, expected 1", len(f.ActorCalls))
	}
}

// ActorCalledN returns true if FakeContext.Actor was called at least n times
func (f *FakeContext) ActorCalledN(n int) bool {
	return len(f.ActorCalls) >= n
}

// AssertActorCalledN calls t.Error if FakeContext.Actor was called less than n times
func (f *FakeContext) AssertActorCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.ActorCalls) < n {
		t.Errorf("FakeContext.Actor called %d times, expected >= %d", len(f.ActorCalls), n)
	}
}

func (_f17 *FakeContext) SetActor(v models.User) {
	invocation := new(SetActorInvocation)

	invocation.Parameters.V = v

	_f17.SetActorHook(v)

	_f17.SetActorCalls = append(_f17.SetActorCalls, invocation)

	return
}

// SetActorCalled returns true if FakeContext.SetActor was called
func (f *FakeContext) SetActorCalled() bool {
	return len(f.SetActorCalls) != 0
}

// AssertSetActorCalled calls t.Error if FakeContext.SetActor was not called
func (f *FakeContext) AssertSetActorCalled(t *testing.T) {
	t.Helper()
	if len(f.SetActorCalls) == 0 {
		t.Error("FakeContext.SetActor not called, expected at least one")
	}
}

// SetActorNotCalled returns true if FakeContext.SetActor was not called
func (f *FakeContext) SetActorNotCalled() bool {
	return len(f.SetActorCalls) == 0
}

// AssertSetActorNotCalled calls t.Error if FakeContext.SetActor was called
func (f *FakeContext) AssertSetActorNotCalled(t *testing.T) {
	t.Helper()
	if len(f.SetActorCalls) != 0 {
		t.Error("FakeContext.SetActor called, expected none")
	}
}

// SetActorCalledOnce returns true if FakeContext.SetActor was called exactly once
func (f *FakeContext) SetActorCalledOnce() bool {
	return len(f.SetActorCalls) == 1
}

// AssertSetActorCalledOnce calls t.Error if FakeContext.SetActor was not called exactly once
func (f *FakeContext) AssertSetActorCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.SetActorCalls) != 1 {
		t.Errorf("FakeContext.SetActor called %d times, expected 1", len(f.SetActorCalls))
	}
}

// SetActorCalledN returns true if FakeContext.SetActor was called at least n times
func (f *FakeContext) SetActorCalledN(n int) bool {
	return len(f.SetActorCalls) >= n
}

// AssertSetActorCalledN calls t.Error if FakeContext.SetActor was called less than n times
func (f *FakeContext) AssertSetActorCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.SetActorCalls) < n {
		t.Errorf("FakeContext.SetActor called %d times, expected >= %d", len(f.SetActorCalls), n)
	}
}

// SetActorCalledWith returns true if FakeContext.SetActor was called with the given values
func (_f18 *FakeContext) SetActorCalledWith(v models.User) (found bool) {
	for _, call := range _f18.SetActorCalls {
		if reflect.DeepEqual(call.Parameters.V, v) {
			found = true
			break
		}
	}

	return
}

// AssertSetActorCalledWith calls t.Error if FakeContext.SetActor was not called with the given values
func (_f19 *FakeContext) AssertSetActorCalledWith(t *testing.T, v models.User) {
	t.Helper()
	var found bool
	for _, call := range _f19.SetActorCalls {
		if reflect.DeepEqual(call.Parameters.V, v) {
			found = true
			break
		}
	}

	if !found {
		t.Error("FakeContext.SetActor not called with expected parameters")
	}
}

// SetActorCalledOnceWith returns true if FakeContext.SetActor was called exactly once with the given values
func (_f20 *FakeContext) SetActorCalledOnceWith(v models.User) bool {
	var count int
	for _, call := range _f20.SetActorCalls {
		if reflect.DeepEqual(call.Parameters.V, v) {
			count++
		}
	}

	return count == 1
}

// AssertSetActorCalledOnceWith calls t.Error if FakeContext.SetActor was not called exactly once with the given values
func (_f21 *FakeContext) AssertSetActorCalledOnceWith(t *testing.T, v models.User) {
	t.Helper()
	var count int
	for _, call := range _f21.SetActorCalls {
		if reflect.DeepEqual(call.Parameters.V, v) {
			count++
		}
	}

	if count != 1 {
		t.Errorf("FakeContext.SetActor called %d times with expected parameters, expected one", count)
	}
}
