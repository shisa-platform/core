// generated by "charlatan -output=./response_charlatan.go Response".  DO NOT EDIT.

package service

import (
	"io"
	"net/http"
	"reflect"
	"testing"
)

// StatusCodeInvocation represents a single call of FakeResponse.StatusCode
type StatusCodeInvocation struct {
	Results struct {
		Ident4 int
	}
}

// HeaderInvocation represents a single call of FakeResponse.Header
type HeaderInvocation struct {
	Results struct {
		Ident5 http.Header
	}
}

// TrailerInvocation represents a single call of FakeResponse.Trailer
type TrailerInvocation struct {
	Results struct {
		Ident6 http.Header
	}
}

// SerializeInvocation represents a single call of FakeResponse.Serialize
type SerializeInvocation struct {
	Parameters struct {
		Ident7 io.Writer
	}
	Results struct {
		Ident8 error
	}
}

/*
FakeResponse is a mock implementation of Response for testing.
Use it in your tests as in this example:

	package example

	func TestWithResponse(t *testing.T) {
		f := &service.FakeResponse{
			StatusCodeHook: func() (ident4 int) {
				// ensure parameters meet expections, signal errors using t, etc
				return
			},
		}

		// test code goes here ...

		// assert state of FakeStatusCode ...
		f.AssertStatusCodeCalledOnce(t)
	}

Create anonymous function implementations for only those interface methods that
should be called in the code under test.  This will force a painc if any
unexpected calls are made to FakeStatusCode.
*/
type FakeResponse struct {
	StatusCodeHook func() int
	HeaderHook     func() http.Header
	TrailerHook    func() http.Header
	SerializeHook  func(io.Writer) error

	StatusCodeCalls []*StatusCodeInvocation
	HeaderCalls     []*HeaderInvocation
	TrailerCalls    []*TrailerInvocation
	SerializeCalls  []*SerializeInvocation
}

// NewFakeResponseDefaultPanic returns an instance of FakeResponse with all hooks configured to panic
func NewFakeResponseDefaultPanic() *FakeResponse {
	return &FakeResponse{
		StatusCodeHook: func() (ident4 int) {
			panic("Unexpected call to Response.StatusCode")
			return
		},
		HeaderHook: func() (ident5 http.Header) {
			panic("Unexpected call to Response.Header")
			return
		},
		TrailerHook: func() (ident6 http.Header) {
			panic("Unexpected call to Response.Trailer")
			return
		},
		SerializeHook: func(io.Writer) (ident8 error) {
			panic("Unexpected call to Response.Serialize")
			return
		},
	}
}

// NewFakeResponseDefaultFatal returns an instance of FakeResponse with all hooks configured to call t.Fatal
func NewFakeResponseDefaultFatal(t *testing.T) *FakeResponse {
	return &FakeResponse{
		StatusCodeHook: func() (ident4 int) {
			t.Fatal("Unexpected call to Response.StatusCode")
			return
		},
		HeaderHook: func() (ident5 http.Header) {
			t.Fatal("Unexpected call to Response.Header")
			return
		},
		TrailerHook: func() (ident6 http.Header) {
			t.Fatal("Unexpected call to Response.Trailer")
			return
		},
		SerializeHook: func(io.Writer) (ident8 error) {
			t.Fatal("Unexpected call to Response.Serialize")
			return
		},
	}
}

// NewFakeResponseDefaultError returns an instance of FakeResponse with all hooks configured to call t.Error
func NewFakeResponseDefaultError(t *testing.T) *FakeResponse {
	return &FakeResponse{
		StatusCodeHook: func() (ident4 int) {
			t.Error("Unexpected call to Response.StatusCode")
			return
		},
		HeaderHook: func() (ident5 http.Header) {
			t.Error("Unexpected call to Response.Header")
			return
		},
		TrailerHook: func() (ident6 http.Header) {
			t.Error("Unexpected call to Response.Trailer")
			return
		},
		SerializeHook: func(io.Writer) (ident8 error) {
			t.Error("Unexpected call to Response.Serialize")
			return
		},
	}
}

func (_f1 *FakeResponse) StatusCode() (ident4 int) {
	invocation := new(StatusCodeInvocation)

	ident4 = _f1.StatusCodeHook()

	invocation.Results.Ident4 = ident4

	_f1.StatusCodeCalls = append(_f1.StatusCodeCalls, invocation)

	return
}

// StatusCodeCalled returns true if FakeResponse.StatusCode was called
func (f *FakeResponse) StatusCodeCalled() bool {
	return len(f.StatusCodeCalls) != 0
}

// AssertStatusCodeCalled calls t.Error if FakeResponse.StatusCode was not called
func (f *FakeResponse) AssertStatusCodeCalled(t *testing.T) {
	t.Helper()
	if len(f.StatusCodeCalls) == 0 {
		t.Error("FakeResponse.StatusCode not called, expected at least one")
	}
}

// StatusCodeNotCalled returns true if FakeResponse.StatusCode was not called
func (f *FakeResponse) StatusCodeNotCalled() bool {
	return len(f.StatusCodeCalls) == 0
}

// AssertStatusCodeNotCalled calls t.Error if FakeResponse.StatusCode was called
func (f *FakeResponse) AssertStatusCodeNotCalled(t *testing.T) {
	t.Helper()
	if len(f.StatusCodeCalls) != 0 {
		t.Error("FakeResponse.StatusCode called, expected none")
	}
}

// StatusCodeCalledOnce returns true if FakeResponse.StatusCode was called exactly once
func (f *FakeResponse) StatusCodeCalledOnce() bool {
	return len(f.StatusCodeCalls) == 1
}

// AssertStatusCodeCalledOnce calls t.Error if FakeResponse.StatusCode was not called exactly once
func (f *FakeResponse) AssertStatusCodeCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.StatusCodeCalls) != 1 {
		t.Errorf("FakeResponse.StatusCode called %d times, expected 1", len(f.StatusCodeCalls))
	}
}

// StatusCodeCalledN returns true if FakeResponse.StatusCode was called at least n times
func (f *FakeResponse) StatusCodeCalledN(n int) bool {
	return len(f.StatusCodeCalls) >= n
}

// AssertStatusCodeCalledN calls t.Error if FakeResponse.StatusCode was called less than n times
func (f *FakeResponse) AssertStatusCodeCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.StatusCodeCalls) < n {
		t.Errorf("FakeResponse.StatusCode called %d times, expected >= %d", len(f.StatusCodeCalls), n)
	}
}

func (_f2 *FakeResponse) Header() (ident5 http.Header) {
	invocation := new(HeaderInvocation)

	ident5 = _f2.HeaderHook()

	invocation.Results.Ident5 = ident5

	_f2.HeaderCalls = append(_f2.HeaderCalls, invocation)

	return
}

// HeaderCalled returns true if FakeResponse.Header was called
func (f *FakeResponse) HeaderCalled() bool {
	return len(f.HeaderCalls) != 0
}

// AssertHeaderCalled calls t.Error if FakeResponse.Header was not called
func (f *FakeResponse) AssertHeaderCalled(t *testing.T) {
	t.Helper()
	if len(f.HeaderCalls) == 0 {
		t.Error("FakeResponse.Header not called, expected at least one")
	}
}

// HeaderNotCalled returns true if FakeResponse.Header was not called
func (f *FakeResponse) HeaderNotCalled() bool {
	return len(f.HeaderCalls) == 0
}

// AssertHeaderNotCalled calls t.Error if FakeResponse.Header was called
func (f *FakeResponse) AssertHeaderNotCalled(t *testing.T) {
	t.Helper()
	if len(f.HeaderCalls) != 0 {
		t.Error("FakeResponse.Header called, expected none")
	}
}

// HeaderCalledOnce returns true if FakeResponse.Header was called exactly once
func (f *FakeResponse) HeaderCalledOnce() bool {
	return len(f.HeaderCalls) == 1
}

// AssertHeaderCalledOnce calls t.Error if FakeResponse.Header was not called exactly once
func (f *FakeResponse) AssertHeaderCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.HeaderCalls) != 1 {
		t.Errorf("FakeResponse.Header called %d times, expected 1", len(f.HeaderCalls))
	}
}

// HeaderCalledN returns true if FakeResponse.Header was called at least n times
func (f *FakeResponse) HeaderCalledN(n int) bool {
	return len(f.HeaderCalls) >= n
}

// AssertHeaderCalledN calls t.Error if FakeResponse.Header was called less than n times
func (f *FakeResponse) AssertHeaderCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.HeaderCalls) < n {
		t.Errorf("FakeResponse.Header called %d times, expected >= %d", len(f.HeaderCalls), n)
	}
}

func (_f3 *FakeResponse) Trailer() (ident6 http.Header) {
	invocation := new(TrailerInvocation)

	ident6 = _f3.TrailerHook()

	invocation.Results.Ident6 = ident6

	_f3.TrailerCalls = append(_f3.TrailerCalls, invocation)

	return
}

// TrailerCalled returns true if FakeResponse.Trailer was called
func (f *FakeResponse) TrailerCalled() bool {
	return len(f.TrailerCalls) != 0
}

// AssertTrailerCalled calls t.Error if FakeResponse.Trailer was not called
func (f *FakeResponse) AssertTrailerCalled(t *testing.T) {
	t.Helper()
	if len(f.TrailerCalls) == 0 {
		t.Error("FakeResponse.Trailer not called, expected at least one")
	}
}

// TrailerNotCalled returns true if FakeResponse.Trailer was not called
func (f *FakeResponse) TrailerNotCalled() bool {
	return len(f.TrailerCalls) == 0
}

// AssertTrailerNotCalled calls t.Error if FakeResponse.Trailer was called
func (f *FakeResponse) AssertTrailerNotCalled(t *testing.T) {
	t.Helper()
	if len(f.TrailerCalls) != 0 {
		t.Error("FakeResponse.Trailer called, expected none")
	}
}

// TrailerCalledOnce returns true if FakeResponse.Trailer was called exactly once
func (f *FakeResponse) TrailerCalledOnce() bool {
	return len(f.TrailerCalls) == 1
}

// AssertTrailerCalledOnce calls t.Error if FakeResponse.Trailer was not called exactly once
func (f *FakeResponse) AssertTrailerCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.TrailerCalls) != 1 {
		t.Errorf("FakeResponse.Trailer called %d times, expected 1", len(f.TrailerCalls))
	}
}

// TrailerCalledN returns true if FakeResponse.Trailer was called at least n times
func (f *FakeResponse) TrailerCalledN(n int) bool {
	return len(f.TrailerCalls) >= n
}

// AssertTrailerCalledN calls t.Error if FakeResponse.Trailer was called less than n times
func (f *FakeResponse) AssertTrailerCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.TrailerCalls) < n {
		t.Errorf("FakeResponse.Trailer called %d times, expected >= %d", len(f.TrailerCalls), n)
	}
}

func (_f4 *FakeResponse) Serialize(ident7 io.Writer) (ident8 error) {
	invocation := new(SerializeInvocation)

	invocation.Parameters.Ident7 = ident7

	ident8 = _f4.SerializeHook(ident7)

	invocation.Results.Ident8 = ident8

	_f4.SerializeCalls = append(_f4.SerializeCalls, invocation)

	return
}

// SerializeCalled returns true if FakeResponse.Serialize was called
func (f *FakeResponse) SerializeCalled() bool {
	return len(f.SerializeCalls) != 0
}

// AssertSerializeCalled calls t.Error if FakeResponse.Serialize was not called
func (f *FakeResponse) AssertSerializeCalled(t *testing.T) {
	t.Helper()
	if len(f.SerializeCalls) == 0 {
		t.Error("FakeResponse.Serialize not called, expected at least one")
	}
}

// SerializeNotCalled returns true if FakeResponse.Serialize was not called
func (f *FakeResponse) SerializeNotCalled() bool {
	return len(f.SerializeCalls) == 0
}

// AssertSerializeNotCalled calls t.Error if FakeResponse.Serialize was called
func (f *FakeResponse) AssertSerializeNotCalled(t *testing.T) {
	t.Helper()
	if len(f.SerializeCalls) != 0 {
		t.Error("FakeResponse.Serialize called, expected none")
	}
}

// SerializeCalledOnce returns true if FakeResponse.Serialize was called exactly once
func (f *FakeResponse) SerializeCalledOnce() bool {
	return len(f.SerializeCalls) == 1
}

// AssertSerializeCalledOnce calls t.Error if FakeResponse.Serialize was not called exactly once
func (f *FakeResponse) AssertSerializeCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.SerializeCalls) != 1 {
		t.Errorf("FakeResponse.Serialize called %d times, expected 1", len(f.SerializeCalls))
	}
}

// SerializeCalledN returns true if FakeResponse.Serialize was called at least n times
func (f *FakeResponse) SerializeCalledN(n int) bool {
	return len(f.SerializeCalls) >= n
}

// AssertSerializeCalledN calls t.Error if FakeResponse.Serialize was called less than n times
func (f *FakeResponse) AssertSerializeCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.SerializeCalls) < n {
		t.Errorf("FakeResponse.Serialize called %d times, expected >= %d", len(f.SerializeCalls), n)
	}
}

// SerializeCalledWith returns true if FakeResponse.Serialize was called with the given values
func (_f5 *FakeResponse) SerializeCalledWith(ident7 io.Writer) (found bool) {
	for _, call := range _f5.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident7, ident7) {
			found = true
			break
		}
	}

	return
}

// AssertSerializeCalledWith calls t.Error if FakeResponse.Serialize was not called with the given values
func (_f6 *FakeResponse) AssertSerializeCalledWith(t *testing.T, ident7 io.Writer) {
	t.Helper()
	var found bool
	for _, call := range _f6.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident7, ident7) {
			found = true
			break
		}
	}

	if !found {
		t.Error("FakeResponse.Serialize not called with expected parameters")
	}
}

// SerializeCalledOnceWith returns true if FakeResponse.Serialize was called exactly once with the given values
func (_f7 *FakeResponse) SerializeCalledOnceWith(ident7 io.Writer) bool {
	var count int
	for _, call := range _f7.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident7, ident7) {
			count++
		}
	}

	return count == 1
}

// AssertSerializeCalledOnceWith calls t.Error if FakeResponse.Serialize was not called exactly once with the given values
func (_f8 *FakeResponse) AssertSerializeCalledOnceWith(t *testing.T, ident7 io.Writer) {
	t.Helper()
	var count int
	for _, call := range _f8.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident7, ident7) {
			count++
		}
	}

	if count != 1 {
		t.Errorf("FakeResponse.Serialize called %d times with expected parameters, expected one", count)
	}
}

// SerializeResultsForCall returns the result values for the first call to FakeResponse.Serialize with the given values
func (_f9 *FakeResponse) SerializeResultsForCall(ident7 io.Writer) (ident8 error, found bool) {
	for _, call := range _f9.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident7, ident7) {
			ident8 = call.Results.Ident8
			found = true
			break
		}
	}

	return
}
