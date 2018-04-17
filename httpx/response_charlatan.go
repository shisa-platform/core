// generated by "charlatan -output=./response_charlatan.go Response".  DO NOT EDIT.

package httpx

import "reflect"

import "github.com/ansel1/merry"

import "net/http"

import "io"

// ResponseStatusCodeInvocation represents a single call of FakeResponse.StatusCode
type ResponseStatusCodeInvocation struct {
	Results struct {
		Ident1 int
	}
}

// ResponseHeadersInvocation represents a single call of FakeResponse.Headers
type ResponseHeadersInvocation struct {
	Results struct {
		Ident1 http.Header
	}
}

// ResponseTrailersInvocation represents a single call of FakeResponse.Trailers
type ResponseTrailersInvocation struct {
	Results struct {
		Ident1 http.Header
	}
}

// ResponseErrInvocation represents a single call of FakeResponse.Err
type ResponseErrInvocation struct {
	Results struct {
		Ident1 error
	}
}

// ResponseSerializeInvocation represents a single call of FakeResponse.Serialize
type ResponseSerializeInvocation struct {
	Parameters struct {
		Ident1 io.Writer
	}
	Results struct {
		Ident2 merry.Error
	}
}

// NewResponseSerializeInvocation creates a new instance of ResponseSerializeInvocation
func NewResponseSerializeInvocation(ident1 io.Writer, ident2 merry.Error) *ResponseSerializeInvocation {
	invocation := new(ResponseSerializeInvocation)

	invocation.Parameters.Ident1 = ident1

	invocation.Results.Ident2 = ident2

	return invocation
}

// ResponseTestingT represents the methods of "testing".T used by charlatan Fakes.  It avoids importing the testing package.
type ResponseTestingT interface {
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Helper()
}

/*
FakeResponse is a mock implementation of Response for testing.
Use it in your tests as in this example:

	package example

	func TestWithResponse(t *testing.T) {
		f := &httpx.FakeResponse{
			StatusCodeHook: func() (ident1 int) {
				// ensure parameters meet expections, signal errors using t, etc
				return
			},
		}

		// test code goes here ...

		// assert state of FakeStatusCode ...
		f.AssertStatusCodeCalledOnce(t)
	}

Create anonymous function implementations for only those interface methods that
should be called in the code under test.  This will force a panic if any
unexpected calls are made to FakeStatusCode.
*/
type FakeResponse struct {
	StatusCodeHook func() int
	HeadersHook    func() http.Header
	TrailersHook   func() http.Header
	ErrHook        func() error
	SerializeHook  func(io.Writer) merry.Error

	StatusCodeCalls []*ResponseStatusCodeInvocation
	HeadersCalls    []*ResponseHeadersInvocation
	TrailersCalls   []*ResponseTrailersInvocation
	ErrCalls        []*ResponseErrInvocation
	SerializeCalls  []*ResponseSerializeInvocation
}

// NewFakeResponseDefaultPanic returns an instance of FakeResponse with all hooks configured to panic
func NewFakeResponseDefaultPanic() *FakeResponse {
	return &FakeResponse{
		StatusCodeHook: func() (ident1 int) {
			panic("Unexpected call to Response.StatusCode")
		},
		HeadersHook: func() (ident1 http.Header) {
			panic("Unexpected call to Response.Headers")
		},
		TrailersHook: func() (ident1 http.Header) {
			panic("Unexpected call to Response.Trailers")
		},
		ErrHook: func() (ident1 error) {
			panic("Unexpected call to Response.Err")
		},
		SerializeHook: func(io.Writer) (ident2 merry.Error) {
			panic("Unexpected call to Response.Serialize")
		},
	}
}

// NewFakeResponseDefaultFatal returns an instance of FakeResponse with all hooks configured to call t.Fatal
func NewFakeResponseDefaultFatal(t ResponseTestingT) *FakeResponse {
	return &FakeResponse{
		StatusCodeHook: func() (ident1 int) {
			t.Fatal("Unexpected call to Response.StatusCode")
			return
		},
		HeadersHook: func() (ident1 http.Header) {
			t.Fatal("Unexpected call to Response.Headers")
			return
		},
		TrailersHook: func() (ident1 http.Header) {
			t.Fatal("Unexpected call to Response.Trailers")
			return
		},
		ErrHook: func() (ident1 error) {
			t.Fatal("Unexpected call to Response.Err")
			return
		},
		SerializeHook: func(io.Writer) (ident2 merry.Error) {
			t.Fatal("Unexpected call to Response.Serialize")
			return
		},
	}
}

// NewFakeResponseDefaultError returns an instance of FakeResponse with all hooks configured to call t.Error
func NewFakeResponseDefaultError(t ResponseTestingT) *FakeResponse {
	return &FakeResponse{
		StatusCodeHook: func() (ident1 int) {
			t.Error("Unexpected call to Response.StatusCode")
			return
		},
		HeadersHook: func() (ident1 http.Header) {
			t.Error("Unexpected call to Response.Headers")
			return
		},
		TrailersHook: func() (ident1 http.Header) {
			t.Error("Unexpected call to Response.Trailers")
			return
		},
		ErrHook: func() (ident1 error) {
			t.Error("Unexpected call to Response.Err")
			return
		},
		SerializeHook: func(io.Writer) (ident2 merry.Error) {
			t.Error("Unexpected call to Response.Serialize")
			return
		},
	}
}

func (f *FakeResponse) Reset() {
	f.StatusCodeCalls = []*ResponseStatusCodeInvocation{}
	f.HeadersCalls = []*ResponseHeadersInvocation{}
	f.TrailersCalls = []*ResponseTrailersInvocation{}
	f.ErrCalls = []*ResponseErrInvocation{}
	f.SerializeCalls = []*ResponseSerializeInvocation{}
}

func (_f1 *FakeResponse) StatusCode() (ident1 int) {
	if _f1.StatusCodeHook == nil {
		panic("Response.StatusCode() called but FakeResponse.StatusCodeHook is nil")
	}

	invocation := new(ResponseStatusCodeInvocation)
	_f1.StatusCodeCalls = append(_f1.StatusCodeCalls, invocation)

	ident1 = _f1.StatusCodeHook()

	invocation.Results.Ident1 = ident1

	return
}

// SetStatusCodeStub configures Response.StatusCode to always return the given values
func (_f2 *FakeResponse) SetStatusCodeStub(ident1 int) {
	_f2.StatusCodeHook = func() int {
		return ident1
	}
}

// StatusCodeCalled returns true if FakeResponse.StatusCode was called
func (f *FakeResponse) StatusCodeCalled() bool {
	return len(f.StatusCodeCalls) != 0
}

// AssertStatusCodeCalled calls t.Error if FakeResponse.StatusCode was not called
func (f *FakeResponse) AssertStatusCodeCalled(t ResponseTestingT) {
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
func (f *FakeResponse) AssertStatusCodeNotCalled(t ResponseTestingT) {
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
func (f *FakeResponse) AssertStatusCodeCalledOnce(t ResponseTestingT) {
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
func (f *FakeResponse) AssertStatusCodeCalledN(t ResponseTestingT, n int) {
	t.Helper()
	if len(f.StatusCodeCalls) < n {
		t.Errorf("FakeResponse.StatusCode called %d times, expected >= %d", len(f.StatusCodeCalls), n)
	}
}

func (_f3 *FakeResponse) Headers() (ident1 http.Header) {
	if _f3.HeadersHook == nil {
		panic("Response.Headers() called but FakeResponse.HeadersHook is nil")
	}

	invocation := new(ResponseHeadersInvocation)
	_f3.HeadersCalls = append(_f3.HeadersCalls, invocation)

	ident1 = _f3.HeadersHook()

	invocation.Results.Ident1 = ident1

	return
}

// SetHeadersStub configures Response.Headers to always return the given values
func (_f4 *FakeResponse) SetHeadersStub(ident1 http.Header) {
	_f4.HeadersHook = func() http.Header {
		return ident1
	}
}

// HeadersCalled returns true if FakeResponse.Headers was called
func (f *FakeResponse) HeadersCalled() bool {
	return len(f.HeadersCalls) != 0
}

// AssertHeadersCalled calls t.Error if FakeResponse.Headers was not called
func (f *FakeResponse) AssertHeadersCalled(t ResponseTestingT) {
	t.Helper()
	if len(f.HeadersCalls) == 0 {
		t.Error("FakeResponse.Headers not called, expected at least one")
	}
}

// HeadersNotCalled returns true if FakeResponse.Headers was not called
func (f *FakeResponse) HeadersNotCalled() bool {
	return len(f.HeadersCalls) == 0
}

// AssertHeadersNotCalled calls t.Error if FakeResponse.Headers was called
func (f *FakeResponse) AssertHeadersNotCalled(t ResponseTestingT) {
	t.Helper()
	if len(f.HeadersCalls) != 0 {
		t.Error("FakeResponse.Headers called, expected none")
	}
}

// HeadersCalledOnce returns true if FakeResponse.Headers was called exactly once
func (f *FakeResponse) HeadersCalledOnce() bool {
	return len(f.HeadersCalls) == 1
}

// AssertHeadersCalledOnce calls t.Error if FakeResponse.Headers was not called exactly once
func (f *FakeResponse) AssertHeadersCalledOnce(t ResponseTestingT) {
	t.Helper()
	if len(f.HeadersCalls) != 1 {
		t.Errorf("FakeResponse.Headers called %d times, expected 1", len(f.HeadersCalls))
	}
}

// HeadersCalledN returns true if FakeResponse.Headers was called at least n times
func (f *FakeResponse) HeadersCalledN(n int) bool {
	return len(f.HeadersCalls) >= n
}

// AssertHeadersCalledN calls t.Error if FakeResponse.Headers was called less than n times
func (f *FakeResponse) AssertHeadersCalledN(t ResponseTestingT, n int) {
	t.Helper()
	if len(f.HeadersCalls) < n {
		t.Errorf("FakeResponse.Headers called %d times, expected >= %d", len(f.HeadersCalls), n)
	}
}

func (_f5 *FakeResponse) Trailers() (ident1 http.Header) {
	if _f5.TrailersHook == nil {
		panic("Response.Trailers() called but FakeResponse.TrailersHook is nil")
	}

	invocation := new(ResponseTrailersInvocation)
	_f5.TrailersCalls = append(_f5.TrailersCalls, invocation)

	ident1 = _f5.TrailersHook()

	invocation.Results.Ident1 = ident1

	return
}

// SetTrailersStub configures Response.Trailers to always return the given values
func (_f6 *FakeResponse) SetTrailersStub(ident1 http.Header) {
	_f6.TrailersHook = func() http.Header {
		return ident1
	}
}

// TrailersCalled returns true if FakeResponse.Trailers was called
func (f *FakeResponse) TrailersCalled() bool {
	return len(f.TrailersCalls) != 0
}

// AssertTrailersCalled calls t.Error if FakeResponse.Trailers was not called
func (f *FakeResponse) AssertTrailersCalled(t ResponseTestingT) {
	t.Helper()
	if len(f.TrailersCalls) == 0 {
		t.Error("FakeResponse.Trailers not called, expected at least one")
	}
}

// TrailersNotCalled returns true if FakeResponse.Trailers was not called
func (f *FakeResponse) TrailersNotCalled() bool {
	return len(f.TrailersCalls) == 0
}

// AssertTrailersNotCalled calls t.Error if FakeResponse.Trailers was called
func (f *FakeResponse) AssertTrailersNotCalled(t ResponseTestingT) {
	t.Helper()
	if len(f.TrailersCalls) != 0 {
		t.Error("FakeResponse.Trailers called, expected none")
	}
}

// TrailersCalledOnce returns true if FakeResponse.Trailers was called exactly once
func (f *FakeResponse) TrailersCalledOnce() bool {
	return len(f.TrailersCalls) == 1
}

// AssertTrailersCalledOnce calls t.Error if FakeResponse.Trailers was not called exactly once
func (f *FakeResponse) AssertTrailersCalledOnce(t ResponseTestingT) {
	t.Helper()
	if len(f.TrailersCalls) != 1 {
		t.Errorf("FakeResponse.Trailers called %d times, expected 1", len(f.TrailersCalls))
	}
}

// TrailersCalledN returns true if FakeResponse.Trailers was called at least n times
func (f *FakeResponse) TrailersCalledN(n int) bool {
	return len(f.TrailersCalls) >= n
}

// AssertTrailersCalledN calls t.Error if FakeResponse.Trailers was called less than n times
func (f *FakeResponse) AssertTrailersCalledN(t ResponseTestingT, n int) {
	t.Helper()
	if len(f.TrailersCalls) < n {
		t.Errorf("FakeResponse.Trailers called %d times, expected >= %d", len(f.TrailersCalls), n)
	}
}

func (_f7 *FakeResponse) Err() (ident1 error) {
	if _f7.ErrHook == nil {
		panic("Response.Err() called but FakeResponse.ErrHook is nil")
	}

	invocation := new(ResponseErrInvocation)
	_f7.ErrCalls = append(_f7.ErrCalls, invocation)

	ident1 = _f7.ErrHook()

	invocation.Results.Ident1 = ident1

	return
}

// SetErrStub configures Response.Err to always return the given values
func (_f8 *FakeResponse) SetErrStub(ident1 error) {
	_f8.ErrHook = func() error {
		return ident1
	}
}

// ErrCalled returns true if FakeResponse.Err was called
func (f *FakeResponse) ErrCalled() bool {
	return len(f.ErrCalls) != 0
}

// AssertErrCalled calls t.Error if FakeResponse.Err was not called
func (f *FakeResponse) AssertErrCalled(t ResponseTestingT) {
	t.Helper()
	if len(f.ErrCalls) == 0 {
		t.Error("FakeResponse.Err not called, expected at least one")
	}
}

// ErrNotCalled returns true if FakeResponse.Err was not called
func (f *FakeResponse) ErrNotCalled() bool {
	return len(f.ErrCalls) == 0
}

// AssertErrNotCalled calls t.Error if FakeResponse.Err was called
func (f *FakeResponse) AssertErrNotCalled(t ResponseTestingT) {
	t.Helper()
	if len(f.ErrCalls) != 0 {
		t.Error("FakeResponse.Err called, expected none")
	}
}

// ErrCalledOnce returns true if FakeResponse.Err was called exactly once
func (f *FakeResponse) ErrCalledOnce() bool {
	return len(f.ErrCalls) == 1
}

// AssertErrCalledOnce calls t.Error if FakeResponse.Err was not called exactly once
func (f *FakeResponse) AssertErrCalledOnce(t ResponseTestingT) {
	t.Helper()
	if len(f.ErrCalls) != 1 {
		t.Errorf("FakeResponse.Err called %d times, expected 1", len(f.ErrCalls))
	}
}

// ErrCalledN returns true if FakeResponse.Err was called at least n times
func (f *FakeResponse) ErrCalledN(n int) bool {
	return len(f.ErrCalls) >= n
}

// AssertErrCalledN calls t.Error if FakeResponse.Err was called less than n times
func (f *FakeResponse) AssertErrCalledN(t ResponseTestingT, n int) {
	t.Helper()
	if len(f.ErrCalls) < n {
		t.Errorf("FakeResponse.Err called %d times, expected >= %d", len(f.ErrCalls), n)
	}
}

func (_f9 *FakeResponse) Serialize(ident1 io.Writer) (ident2 merry.Error) {
	if _f9.SerializeHook == nil {
		panic("Response.Serialize() called but FakeResponse.SerializeHook is nil")
	}

	invocation := new(ResponseSerializeInvocation)
	_f9.SerializeCalls = append(_f9.SerializeCalls, invocation)

	invocation.Parameters.Ident1 = ident1

	ident2 = _f9.SerializeHook(ident1)

	invocation.Results.Ident2 = ident2

	return
}

// SetSerializeStub configures Response.Serialize to always return the given values
func (_f10 *FakeResponse) SetSerializeStub(ident2 merry.Error) {
	_f10.SerializeHook = func(io.Writer) merry.Error {
		return ident2
	}
}

// SetSerializeInvocation configures Response.Serialize to return the given results when called with the given parameters
// If no match is found for an invocation the result(s) of the fallback function are returned
func (_f11 *FakeResponse) SetSerializeInvocation(calls_f12 []*ResponseSerializeInvocation, fallback_f13 func() merry.Error) {
	_f11.SerializeHook = func(ident1 io.Writer) (ident2 merry.Error) {
		for _, call := range calls_f12 {
			if reflect.DeepEqual(call.Parameters.Ident1, ident1) {
				ident2 = call.Results.Ident2

				return
			}
		}

		return fallback_f13()
	}
}

// SerializeCalled returns true if FakeResponse.Serialize was called
func (f *FakeResponse) SerializeCalled() bool {
	return len(f.SerializeCalls) != 0
}

// AssertSerializeCalled calls t.Error if FakeResponse.Serialize was not called
func (f *FakeResponse) AssertSerializeCalled(t ResponseTestingT) {
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
func (f *FakeResponse) AssertSerializeNotCalled(t ResponseTestingT) {
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
func (f *FakeResponse) AssertSerializeCalledOnce(t ResponseTestingT) {
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
func (f *FakeResponse) AssertSerializeCalledN(t ResponseTestingT, n int) {
	t.Helper()
	if len(f.SerializeCalls) < n {
		t.Errorf("FakeResponse.Serialize called %d times, expected >= %d", len(f.SerializeCalls), n)
	}
}

// SerializeCalledWith returns true if FakeResponse.Serialize was called with the given values
func (_f14 *FakeResponse) SerializeCalledWith(ident1 io.Writer) (found bool) {
	for _, call := range _f14.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) {
			found = true
			break
		}
	}

	return
}

// AssertSerializeCalledWith calls t.Error if FakeResponse.Serialize was not called with the given values
func (_f15 *FakeResponse) AssertSerializeCalledWith(t ResponseTestingT, ident1 io.Writer) {
	t.Helper()
	var found bool
	for _, call := range _f15.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) {
			found = true
			break
		}
	}

	if !found {
		t.Error("FakeResponse.Serialize not called with expected parameters")
	}
}

// SerializeCalledOnceWith returns true if FakeResponse.Serialize was called exactly once with the given values
func (_f16 *FakeResponse) SerializeCalledOnceWith(ident1 io.Writer) bool {
	var count int
	for _, call := range _f16.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) {
			count++
		}
	}

	return count == 1
}

// AssertSerializeCalledOnceWith calls t.Error if FakeResponse.Serialize was not called exactly once with the given values
func (_f17 *FakeResponse) AssertSerializeCalledOnceWith(t ResponseTestingT, ident1 io.Writer) {
	t.Helper()
	var count int
	for _, call := range _f17.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) {
			count++
		}
	}

	if count != 1 {
		t.Errorf("FakeResponse.Serialize called %d times with expected parameters, expected one", count)
	}
}

// SerializeResultsForCall returns the result values for the first call to FakeResponse.Serialize with the given values
func (_f18 *FakeResponse) SerializeResultsForCall(ident1 io.Writer) (ident2 merry.Error, found bool) {
	for _, call := range _f18.SerializeCalls {
		if reflect.DeepEqual(call.Parameters.Ident1, ident1) {
			ident2 = call.Results.Ident2
			found = true
			break
		}
	}

	return
}
