// generated by "charlatan -output=./consulkvgetter_charlatan.go KVGetter".  DO NOT EDIT.

package env

import consulapi "github.com/hashicorp/consul/api"
import "reflect"

// KVGetterGetInvocation represents a single call of FakeKVGetter.Get
type KVGetterGetInvocation struct {
	Parameters struct {
		Ident1 string
		Ident2 *consulapi.QueryOptions
	}
	Results struct {
		Ident3 *consulapi.KVPair
		Ident4 *consulapi.QueryMeta
		Ident5 error
	}
}

// KVGetterTestingT represents the methods of "testing".T used by charlatan Fakes.  It avoids importing the testing package.
type KVGetterTestingT interface {
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Helper()
}

/*
FakeKVGetter is a mock implementation of KVGetter for testing.
Use it in your tests as in this example:

	package example

	func TestWithKVGetter(t *testing.T) {
		f := &env.FakeKVGetter{
			GetHook: func(ident1 string, ident2 *consulapi.QueryOptions) (ident3 *consulapi.KVPair, ident4 *consulapi.QueryMeta, ident5 error) {
				// ensure parameters meet expections, signal errors using t, etc
				return
			},
		}

		// test code goes here ...

		// assert state of FakeGet ...
		f.AssertGetCalledOnce(t)
	}

Create anonymous function implementations for only those interface methods that
should be called in the code under test.  This will force a panic if any
unexpected calls are made to FakeGet.
*/
type FakeKVGetter struct {
	GetHook func(string, *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error)

	GetCalls []*KVGetterGetInvocation
}

// NewFakeKVGetterDefaultPanic returns an instance of FakeKVGetter with all hooks configured to panic
func NewFakeKVGetterDefaultPanic() *FakeKVGetter {
	return &FakeKVGetter{
		GetHook: func(string, *consulapi.QueryOptions) (ident3 *consulapi.KVPair, ident4 *consulapi.QueryMeta, ident5 error) {
			panic("Unexpected call to KVGetter.Get")
		},
	}
}

// NewFakeKVGetterDefaultFatal returns an instance of FakeKVGetter with all hooks configured to call t.Fatal
func NewFakeKVGetterDefaultFatal(t KVGetterTestingT) *FakeKVGetter {
	return &FakeKVGetter{
		GetHook: func(string, *consulapi.QueryOptions) (ident3 *consulapi.KVPair, ident4 *consulapi.QueryMeta, ident5 error) {
			t.Fatal("Unexpected call to KVGetter.Get")
			return
		},
	}
}

// NewFakeKVGetterDefaultError returns an instance of FakeKVGetter with all hooks configured to call t.Error
func NewFakeKVGetterDefaultError(t KVGetterTestingT) *FakeKVGetter {
	return &FakeKVGetter{
		GetHook: func(string, *consulapi.QueryOptions) (ident3 *consulapi.KVPair, ident4 *consulapi.QueryMeta, ident5 error) {
			t.Error("Unexpected call to KVGetter.Get")
			return
		},
	}
}

func (f *FakeKVGetter) Reset() {
	f.GetCalls = []*KVGetterGetInvocation{}
}

func (_f1 *FakeKVGetter) Get(ident1 string, ident2 *consulapi.QueryOptions) (ident3 *consulapi.KVPair, ident4 *consulapi.QueryMeta, ident5 error) {
	invocation := new(KVGetterGetInvocation)

	invocation.Parameters.Ident1 = ident1
	invocation.Parameters.Ident2 = ident2

	ident3, ident4, ident5 = _f1.GetHook(ident1, ident2)

	invocation.Results.Ident3 = ident3
	invocation.Results.Ident4 = ident4
	invocation.Results.Ident5 = ident5

	_f1.GetCalls = append(_f1.GetCalls, invocation)

	return
}

// GetCalled returns true if FakeKVGetter.Get was called
func (f *FakeKVGetter) GetCalled() bool {
	return len(f.GetCalls) != 0
}

// AssertGetCalled calls t.Error if FakeKVGetter.Get was not called
func (f *FakeKVGetter) AssertGetCalled(t KVGetterTestingT) {
	t.Helper()
	if len(f.GetCalls) == 0 {
		t.Error("FakeKVGetter.Get not called, expected at least one")
	}
}

// GetNotCalled returns true if FakeKVGetter.Get was not called
func (f *FakeKVGetter) GetNotCalled() bool {
	return len(f.GetCalls) == 0
}

// AssertGetNotCalled calls t.Error if FakeKVGetter.Get was called
func (f *FakeKVGetter) AssertGetNotCalled(t KVGetterTestingT) {
	t.Helper()
	if len(f.GetCalls) != 0 {
		t.Error("FakeKVGetter.Get called, expected none")
	}
}

// GetCalledOnce returns true if FakeKVGetter.Get was called exactly once
func (f *FakeKVGetter) GetCalledOnce() bool {
	return len(f.GetCalls) == 1
}

// AssertGetCalledOnce calls t.Error if FakeKVGetter.Get was not called exactly once
func (f *FakeKVGetter) AssertGetCalledOnce(t KVGetterTestingT) {
	t.Helper()
	if len(f.GetCalls) != 1 {
		t.Errorf("FakeKVGetter.Get called %d times, expected 1", len(f.GetCalls))
	}
}

// GetCalledN returns true if FakeKVGetter.Get was called at least n times
func (f *FakeKVGetter) GetCalledN(n int) bool {
	return len(f.GetCalls) >= n
}

// AssertGetCalledN calls t.Error if FakeKVGetter.Get was called less than n times
func (f *FakeKVGetter) AssertGetCalledN(t KVGetterTestingT, n int) {
	t.Helper()
	if len(f.GetCalls) < n {
		t.Errorf("FakeKVGetter.Get called %d times, expected >= %d", len(f.GetCalls), n)
	}
}

// GetCalledWith returns true if FakeKVGetter.Get was called with the given values
func (_f2 *FakeKVGetter) GetCalledWith(ident1 string, ident2 *consulapi.QueryOptions) (found bool) {
	for _, call := range _f2.GetCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) && reflect.DeepEqual(call.Parameters.Ident2, ident2) {
			found = true
			break
		}
	}

	return
}

// AssertGetCalledWith calls t.Error if FakeKVGetter.Get was not called with the given values
func (_f3 *FakeKVGetter) AssertGetCalledWith(t KVGetterTestingT, ident1 string, ident2 *consulapi.QueryOptions) {
	t.Helper()
	var found bool
	for _, call := range _f3.GetCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) && reflect.DeepEqual(call.Parameters.Ident2, ident2) {
			found = true
			break
		}
	}

	if !found {
		t.Error("FakeKVGetter.Get not called with expected parameters")
	}
}

// GetCalledOnceWith returns true if FakeKVGetter.Get was called exactly once with the given values
func (_f4 *FakeKVGetter) GetCalledOnceWith(ident1 string, ident2 *consulapi.QueryOptions) bool {
	var count int
	for _, call := range _f4.GetCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) && reflect.DeepEqual(call.Parameters.Ident2, ident2) {
			count++
		}
	}

	return count == 1
}

// AssertGetCalledOnceWith calls t.Error if FakeKVGetter.Get was not called exactly once with the given values
func (_f5 *FakeKVGetter) AssertGetCalledOnceWith(t KVGetterTestingT, ident1 string, ident2 *consulapi.QueryOptions) {
	t.Helper()
	var count int
	for _, call := range _f5.GetCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) && reflect.DeepEqual(call.Parameters.Ident2, ident2) {
			count++
		}
	}

	if count != 1 {
		t.Errorf("FakeKVGetter.Get called %d times with expected parameters, expected one", count)
	}
}

// GetResultsForCall returns the result values for the first call to FakeKVGetter.Get with the given values
func (_f6 *FakeKVGetter) GetResultsForCall(ident1 string, ident2 *consulapi.QueryOptions) (ident3 *consulapi.KVPair, ident4 *consulapi.QueryMeta, ident5 error, found bool) {
	for _, call := range _f6.GetCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) && reflect.DeepEqual(call.Parameters.Ident2, ident2) {
			ident3 = call.Results.Ident3
			ident4 = call.Results.Ident4
			ident5 = call.Results.Ident5
			found = true
			break
		}
	}

	return
}
