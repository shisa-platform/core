// generated by "charlatan -output=./service_charlatan.go Service".  DO NOT EDIT.

package service

import (
	"testing"
)

// ServiceNameInvocation represents a single call of FakeService.Name
type ServiceNameInvocation struct {
	Results struct {
		Ident1 string
	}
}

// ServiceEndpointsInvocation represents a single call of FakeService.Endpoints
type ServiceEndpointsInvocation struct {
	Results struct {
		Ident1 []Endpoint
	}
}

// ServiceHandlersInvocation represents a single call of FakeService.Handlers
type ServiceHandlersInvocation struct {
	Results struct {
		Ident1 []Handler
	}
}

// ServiceMalformedQueryParameterHandlerInvocation represents a single call of FakeService.MalformedQueryParameterHandler
type ServiceMalformedQueryParameterHandlerInvocation struct {
	Results struct {
		Ident1 Handler
	}
}

// ServiceMethodNotAllowedHandlerInvocation represents a single call of FakeService.MethodNotAllowedHandler
type ServiceMethodNotAllowedHandlerInvocation struct {
	Results struct {
		Ident1 Handler
	}
}

// ServiceRedirectHandlerInvocation represents a single call of FakeService.RedirectHandler
type ServiceRedirectHandlerInvocation struct {
	Results struct {
		Ident1 Handler
	}
}

// ServiceInternalServerErrorHandlerInvocation represents a single call of FakeService.InternalServerErrorHandler
type ServiceInternalServerErrorHandlerInvocation struct {
	Results struct {
		Ident1 ErrorHandler
	}
}

/*
FakeService is a mock implementation of Service for testing.
Use it in your tests as in this example:

	package example

	func TestWithService(t *testing.T) {
		f := &service.FakeService{
			NameHook: func() (ident1 string) {
				// ensure parameters meet expections, signal errors using t, etc
				return
			},
		}

		// test code goes here ...

		// assert state of FakeName ...
		f.AssertNameCalledOnce(t)
	}

Create anonymous function implementations for only those interface methods that
should be called in the code under test.  This will force a panic if any
unexpected calls are made to FakeName.
*/
type FakeService struct {
	NameHook                           func() string
	EndpointsHook                      func() []Endpoint
	HandlersHook                       func() []Handler
	MalformedQueryParameterHandlerHook func() Handler
	MethodNotAllowedHandlerHook        func() Handler
	RedirectHandlerHook                func() Handler
	InternalServerErrorHandlerHook     func() ErrorHandler

	NameCalls                           []*ServiceNameInvocation
	EndpointsCalls                      []*ServiceEndpointsInvocation
	HandlersCalls                       []*ServiceHandlersInvocation
	MalformedQueryParameterHandlerCalls []*ServiceMalformedQueryParameterHandlerInvocation
	MethodNotAllowedHandlerCalls        []*ServiceMethodNotAllowedHandlerInvocation
	RedirectHandlerCalls                []*ServiceRedirectHandlerInvocation
	InternalServerErrorHandlerCalls     []*ServiceInternalServerErrorHandlerInvocation
}

// NewFakeServiceDefaultPanic returns an instance of FakeService with all hooks configured to panic
func NewFakeServiceDefaultPanic() *FakeService {
	return &FakeService{
		NameHook: func() (ident1 string) {
			panic("Unexpected call to Service.Name")
		},
		EndpointsHook: func() (ident1 []Endpoint) {
			panic("Unexpected call to Service.Endpoints")
		},
		HandlersHook: func() (ident1 []Handler) {
			panic("Unexpected call to Service.Handlers")
		},
		MalformedQueryParameterHandlerHook: func() (ident1 Handler) {
			panic("Unexpected call to Service.MalformedQueryParameterHandler")
		},
		MethodNotAllowedHandlerHook: func() (ident1 Handler) {
			panic("Unexpected call to Service.MethodNotAllowedHandler")
		},
		RedirectHandlerHook: func() (ident1 Handler) {
			panic("Unexpected call to Service.RedirectHandler")
		},
		InternalServerErrorHandlerHook: func() (ident1 ErrorHandler) {
			panic("Unexpected call to Service.InternalServerErrorHandler")
		},
	}
}

// NewFakeServiceDefaultFatal returns an instance of FakeService with all hooks configured to call t.Fatal
func NewFakeServiceDefaultFatal(t *testing.T) *FakeService {
	return &FakeService{
		NameHook: func() (ident1 string) {
			t.Fatal("Unexpected call to Service.Name")
			return
		},
		EndpointsHook: func() (ident1 []Endpoint) {
			t.Fatal("Unexpected call to Service.Endpoints")
			return
		},
		HandlersHook: func() (ident1 []Handler) {
			t.Fatal("Unexpected call to Service.Handlers")
			return
		},
		MalformedQueryParameterHandlerHook: func() (ident1 Handler) {
			t.Fatal("Unexpected call to Service.MalformedQueryParameterHandler")
			return
		},
		MethodNotAllowedHandlerHook: func() (ident1 Handler) {
			t.Fatal("Unexpected call to Service.MethodNotAllowedHandler")
			return
		},
		RedirectHandlerHook: func() (ident1 Handler) {
			t.Fatal("Unexpected call to Service.RedirectHandler")
			return
		},
		InternalServerErrorHandlerHook: func() (ident1 ErrorHandler) {
			t.Fatal("Unexpected call to Service.InternalServerErrorHandler")
			return
		},
	}
}

// NewFakeServiceDefaultError returns an instance of FakeService with all hooks configured to call t.Error
func NewFakeServiceDefaultError(t *testing.T) *FakeService {
	return &FakeService{
		NameHook: func() (ident1 string) {
			t.Error("Unexpected call to Service.Name")
			return
		},
		EndpointsHook: func() (ident1 []Endpoint) {
			t.Error("Unexpected call to Service.Endpoints")
			return
		},
		HandlersHook: func() (ident1 []Handler) {
			t.Error("Unexpected call to Service.Handlers")
			return
		},
		MalformedQueryParameterHandlerHook: func() (ident1 Handler) {
			t.Error("Unexpected call to Service.MalformedQueryParameterHandler")
			return
		},
		MethodNotAllowedHandlerHook: func() (ident1 Handler) {
			t.Error("Unexpected call to Service.MethodNotAllowedHandler")
			return
		},
		RedirectHandlerHook: func() (ident1 Handler) {
			t.Error("Unexpected call to Service.RedirectHandler")
			return
		},
		InternalServerErrorHandlerHook: func() (ident1 ErrorHandler) {
			t.Error("Unexpected call to Service.InternalServerErrorHandler")
			return
		},
	}
}

func (f *FakeService) Reset() {
	f.NameCalls = []*ServiceNameInvocation{}
	f.EndpointsCalls = []*ServiceEndpointsInvocation{}
	f.HandlersCalls = []*ServiceHandlersInvocation{}
	f.MalformedQueryParameterHandlerCalls = []*ServiceMalformedQueryParameterHandlerInvocation{}
	f.MethodNotAllowedHandlerCalls = []*ServiceMethodNotAllowedHandlerInvocation{}
	f.RedirectHandlerCalls = []*ServiceRedirectHandlerInvocation{}
	f.InternalServerErrorHandlerCalls = []*ServiceInternalServerErrorHandlerInvocation{}
}

func (_f1 *FakeService) Name() (ident1 string) {
	invocation := new(ServiceNameInvocation)

	ident1 = _f1.NameHook()

	invocation.Results.Ident1 = ident1

	_f1.NameCalls = append(_f1.NameCalls, invocation)

	return
}

// NameCalled returns true if FakeService.Name was called
func (f *FakeService) NameCalled() bool {
	return len(f.NameCalls) != 0
}

// AssertNameCalled calls t.Error if FakeService.Name was not called
func (f *FakeService) AssertNameCalled(t *testing.T) {
	t.Helper()
	if len(f.NameCalls) == 0 {
		t.Error("FakeService.Name not called, expected at least one")
	}
}

// NameNotCalled returns true if FakeService.Name was not called
func (f *FakeService) NameNotCalled() bool {
	return len(f.NameCalls) == 0
}

// AssertNameNotCalled calls t.Error if FakeService.Name was called
func (f *FakeService) AssertNameNotCalled(t *testing.T) {
	t.Helper()
	if len(f.NameCalls) != 0 {
		t.Error("FakeService.Name called, expected none")
	}
}

// NameCalledOnce returns true if FakeService.Name was called exactly once
func (f *FakeService) NameCalledOnce() bool {
	return len(f.NameCalls) == 1
}

// AssertNameCalledOnce calls t.Error if FakeService.Name was not called exactly once
func (f *FakeService) AssertNameCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.NameCalls) != 1 {
		t.Errorf("FakeService.Name called %d times, expected 1", len(f.NameCalls))
	}
}

// NameCalledN returns true if FakeService.Name was called at least n times
func (f *FakeService) NameCalledN(n int) bool {
	return len(f.NameCalls) >= n
}

// AssertNameCalledN calls t.Error if FakeService.Name was called less than n times
func (f *FakeService) AssertNameCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.NameCalls) < n {
		t.Errorf("FakeService.Name called %d times, expected >= %d", len(f.NameCalls), n)
	}
}

func (_f2 *FakeService) Endpoints() (ident1 []Endpoint) {
	invocation := new(ServiceEndpointsInvocation)

	ident1 = _f2.EndpointsHook()

	invocation.Results.Ident1 = ident1

	_f2.EndpointsCalls = append(_f2.EndpointsCalls, invocation)

	return
}

// EndpointsCalled returns true if FakeService.Endpoints was called
func (f *FakeService) EndpointsCalled() bool {
	return len(f.EndpointsCalls) != 0
}

// AssertEndpointsCalled calls t.Error if FakeService.Endpoints was not called
func (f *FakeService) AssertEndpointsCalled(t *testing.T) {
	t.Helper()
	if len(f.EndpointsCalls) == 0 {
		t.Error("FakeService.Endpoints not called, expected at least one")
	}
}

// EndpointsNotCalled returns true if FakeService.Endpoints was not called
func (f *FakeService) EndpointsNotCalled() bool {
	return len(f.EndpointsCalls) == 0
}

// AssertEndpointsNotCalled calls t.Error if FakeService.Endpoints was called
func (f *FakeService) AssertEndpointsNotCalled(t *testing.T) {
	t.Helper()
	if len(f.EndpointsCalls) != 0 {
		t.Error("FakeService.Endpoints called, expected none")
	}
}

// EndpointsCalledOnce returns true if FakeService.Endpoints was called exactly once
func (f *FakeService) EndpointsCalledOnce() bool {
	return len(f.EndpointsCalls) == 1
}

// AssertEndpointsCalledOnce calls t.Error if FakeService.Endpoints was not called exactly once
func (f *FakeService) AssertEndpointsCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.EndpointsCalls) != 1 {
		t.Errorf("FakeService.Endpoints called %d times, expected 1", len(f.EndpointsCalls))
	}
}

// EndpointsCalledN returns true if FakeService.Endpoints was called at least n times
func (f *FakeService) EndpointsCalledN(n int) bool {
	return len(f.EndpointsCalls) >= n
}

// AssertEndpointsCalledN calls t.Error if FakeService.Endpoints was called less than n times
func (f *FakeService) AssertEndpointsCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.EndpointsCalls) < n {
		t.Errorf("FakeService.Endpoints called %d times, expected >= %d", len(f.EndpointsCalls), n)
	}
}

func (_f3 *FakeService) Handlers() (ident1 []Handler) {
	invocation := new(ServiceHandlersInvocation)

	ident1 = _f3.HandlersHook()

	invocation.Results.Ident1 = ident1

	_f3.HandlersCalls = append(_f3.HandlersCalls, invocation)

	return
}

// HandlersCalled returns true if FakeService.Handlers was called
func (f *FakeService) HandlersCalled() bool {
	return len(f.HandlersCalls) != 0
}

// AssertHandlersCalled calls t.Error if FakeService.Handlers was not called
func (f *FakeService) AssertHandlersCalled(t *testing.T) {
	t.Helper()
	if len(f.HandlersCalls) == 0 {
		t.Error("FakeService.Handlers not called, expected at least one")
	}
}

// HandlersNotCalled returns true if FakeService.Handlers was not called
func (f *FakeService) HandlersNotCalled() bool {
	return len(f.HandlersCalls) == 0
}

// AssertHandlersNotCalled calls t.Error if FakeService.Handlers was called
func (f *FakeService) AssertHandlersNotCalled(t *testing.T) {
	t.Helper()
	if len(f.HandlersCalls) != 0 {
		t.Error("FakeService.Handlers called, expected none")
	}
}

// HandlersCalledOnce returns true if FakeService.Handlers was called exactly once
func (f *FakeService) HandlersCalledOnce() bool {
	return len(f.HandlersCalls) == 1
}

// AssertHandlersCalledOnce calls t.Error if FakeService.Handlers was not called exactly once
func (f *FakeService) AssertHandlersCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.HandlersCalls) != 1 {
		t.Errorf("FakeService.Handlers called %d times, expected 1", len(f.HandlersCalls))
	}
}

// HandlersCalledN returns true if FakeService.Handlers was called at least n times
func (f *FakeService) HandlersCalledN(n int) bool {
	return len(f.HandlersCalls) >= n
}

// AssertHandlersCalledN calls t.Error if FakeService.Handlers was called less than n times
func (f *FakeService) AssertHandlersCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.HandlersCalls) < n {
		t.Errorf("FakeService.Handlers called %d times, expected >= %d", len(f.HandlersCalls), n)
	}
}

func (_f4 *FakeService) MalformedQueryParameterHandler() (ident1 Handler) {
	invocation := new(ServiceMalformedQueryParameterHandlerInvocation)

	ident1 = _f4.MalformedQueryParameterHandlerHook()

	invocation.Results.Ident1 = ident1

	_f4.MalformedQueryParameterHandlerCalls = append(_f4.MalformedQueryParameterHandlerCalls, invocation)

	return
}

// MalformedQueryParameterHandlerCalled returns true if FakeService.MalformedQueryParameterHandler was called
func (f *FakeService) MalformedQueryParameterHandlerCalled() bool {
	return len(f.MalformedQueryParameterHandlerCalls) != 0
}

// AssertMalformedQueryParameterHandlerCalled calls t.Error if FakeService.MalformedQueryParameterHandler was not called
func (f *FakeService) AssertMalformedQueryParameterHandlerCalled(t *testing.T) {
	t.Helper()
	if len(f.MalformedQueryParameterHandlerCalls) == 0 {
		t.Error("FakeService.MalformedQueryParameterHandler not called, expected at least one")
	}
}

// MalformedQueryParameterHandlerNotCalled returns true if FakeService.MalformedQueryParameterHandler was not called
func (f *FakeService) MalformedQueryParameterHandlerNotCalled() bool {
	return len(f.MalformedQueryParameterHandlerCalls) == 0
}

// AssertMalformedQueryParameterHandlerNotCalled calls t.Error if FakeService.MalformedQueryParameterHandler was called
func (f *FakeService) AssertMalformedQueryParameterHandlerNotCalled(t *testing.T) {
	t.Helper()
	if len(f.MalformedQueryParameterHandlerCalls) != 0 {
		t.Error("FakeService.MalformedQueryParameterHandler called, expected none")
	}
}

// MalformedQueryParameterHandlerCalledOnce returns true if FakeService.MalformedQueryParameterHandler was called exactly once
func (f *FakeService) MalformedQueryParameterHandlerCalledOnce() bool {
	return len(f.MalformedQueryParameterHandlerCalls) == 1
}

// AssertMalformedQueryParameterHandlerCalledOnce calls t.Error if FakeService.MalformedQueryParameterHandler was not called exactly once
func (f *FakeService) AssertMalformedQueryParameterHandlerCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.MalformedQueryParameterHandlerCalls) != 1 {
		t.Errorf("FakeService.MalformedQueryParameterHandler called %d times, expected 1", len(f.MalformedQueryParameterHandlerCalls))
	}
}

// MalformedQueryParameterHandlerCalledN returns true if FakeService.MalformedQueryParameterHandler was called at least n times
func (f *FakeService) MalformedQueryParameterHandlerCalledN(n int) bool {
	return len(f.MalformedQueryParameterHandlerCalls) >= n
}

// AssertMalformedQueryParameterHandlerCalledN calls t.Error if FakeService.MalformedQueryParameterHandler was called less than n times
func (f *FakeService) AssertMalformedQueryParameterHandlerCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.MalformedQueryParameterHandlerCalls) < n {
		t.Errorf("FakeService.MalformedQueryParameterHandler called %d times, expected >= %d", len(f.MalformedQueryParameterHandlerCalls), n)
	}
}

func (_f5 *FakeService) MethodNotAllowedHandler() (ident1 Handler) {
	invocation := new(ServiceMethodNotAllowedHandlerInvocation)

	ident1 = _f5.MethodNotAllowedHandlerHook()

	invocation.Results.Ident1 = ident1

	_f5.MethodNotAllowedHandlerCalls = append(_f5.MethodNotAllowedHandlerCalls, invocation)

	return
}

// MethodNotAllowedHandlerCalled returns true if FakeService.MethodNotAllowedHandler was called
func (f *FakeService) MethodNotAllowedHandlerCalled() bool {
	return len(f.MethodNotAllowedHandlerCalls) != 0
}

// AssertMethodNotAllowedHandlerCalled calls t.Error if FakeService.MethodNotAllowedHandler was not called
func (f *FakeService) AssertMethodNotAllowedHandlerCalled(t *testing.T) {
	t.Helper()
	if len(f.MethodNotAllowedHandlerCalls) == 0 {
		t.Error("FakeService.MethodNotAllowedHandler not called, expected at least one")
	}
}

// MethodNotAllowedHandlerNotCalled returns true if FakeService.MethodNotAllowedHandler was not called
func (f *FakeService) MethodNotAllowedHandlerNotCalled() bool {
	return len(f.MethodNotAllowedHandlerCalls) == 0
}

// AssertMethodNotAllowedHandlerNotCalled calls t.Error if FakeService.MethodNotAllowedHandler was called
func (f *FakeService) AssertMethodNotAllowedHandlerNotCalled(t *testing.T) {
	t.Helper()
	if len(f.MethodNotAllowedHandlerCalls) != 0 {
		t.Error("FakeService.MethodNotAllowedHandler called, expected none")
	}
}

// MethodNotAllowedHandlerCalledOnce returns true if FakeService.MethodNotAllowedHandler was called exactly once
func (f *FakeService) MethodNotAllowedHandlerCalledOnce() bool {
	return len(f.MethodNotAllowedHandlerCalls) == 1
}

// AssertMethodNotAllowedHandlerCalledOnce calls t.Error if FakeService.MethodNotAllowedHandler was not called exactly once
func (f *FakeService) AssertMethodNotAllowedHandlerCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.MethodNotAllowedHandlerCalls) != 1 {
		t.Errorf("FakeService.MethodNotAllowedHandler called %d times, expected 1", len(f.MethodNotAllowedHandlerCalls))
	}
}

// MethodNotAllowedHandlerCalledN returns true if FakeService.MethodNotAllowedHandler was called at least n times
func (f *FakeService) MethodNotAllowedHandlerCalledN(n int) bool {
	return len(f.MethodNotAllowedHandlerCalls) >= n
}

// AssertMethodNotAllowedHandlerCalledN calls t.Error if FakeService.MethodNotAllowedHandler was called less than n times
func (f *FakeService) AssertMethodNotAllowedHandlerCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.MethodNotAllowedHandlerCalls) < n {
		t.Errorf("FakeService.MethodNotAllowedHandler called %d times, expected >= %d", len(f.MethodNotAllowedHandlerCalls), n)
	}
}

func (_f6 *FakeService) RedirectHandler() (ident1 Handler) {
	invocation := new(ServiceRedirectHandlerInvocation)

	ident1 = _f6.RedirectHandlerHook()

	invocation.Results.Ident1 = ident1

	_f6.RedirectHandlerCalls = append(_f6.RedirectHandlerCalls, invocation)

	return
}

// RedirectHandlerCalled returns true if FakeService.RedirectHandler was called
func (f *FakeService) RedirectHandlerCalled() bool {
	return len(f.RedirectHandlerCalls) != 0
}

// AssertRedirectHandlerCalled calls t.Error if FakeService.RedirectHandler was not called
func (f *FakeService) AssertRedirectHandlerCalled(t *testing.T) {
	t.Helper()
	if len(f.RedirectHandlerCalls) == 0 {
		t.Error("FakeService.RedirectHandler not called, expected at least one")
	}
}

// RedirectHandlerNotCalled returns true if FakeService.RedirectHandler was not called
func (f *FakeService) RedirectHandlerNotCalled() bool {
	return len(f.RedirectHandlerCalls) == 0
}

// AssertRedirectHandlerNotCalled calls t.Error if FakeService.RedirectHandler was called
func (f *FakeService) AssertRedirectHandlerNotCalled(t *testing.T) {
	t.Helper()
	if len(f.RedirectHandlerCalls) != 0 {
		t.Error("FakeService.RedirectHandler called, expected none")
	}
}

// RedirectHandlerCalledOnce returns true if FakeService.RedirectHandler was called exactly once
func (f *FakeService) RedirectHandlerCalledOnce() bool {
	return len(f.RedirectHandlerCalls) == 1
}

// AssertRedirectHandlerCalledOnce calls t.Error if FakeService.RedirectHandler was not called exactly once
func (f *FakeService) AssertRedirectHandlerCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.RedirectHandlerCalls) != 1 {
		t.Errorf("FakeService.RedirectHandler called %d times, expected 1", len(f.RedirectHandlerCalls))
	}
}

// RedirectHandlerCalledN returns true if FakeService.RedirectHandler was called at least n times
func (f *FakeService) RedirectHandlerCalledN(n int) bool {
	return len(f.RedirectHandlerCalls) >= n
}

// AssertRedirectHandlerCalledN calls t.Error if FakeService.RedirectHandler was called less than n times
func (f *FakeService) AssertRedirectHandlerCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.RedirectHandlerCalls) < n {
		t.Errorf("FakeService.RedirectHandler called %d times, expected >= %d", len(f.RedirectHandlerCalls), n)
	}
}

func (_f7 *FakeService) InternalServerErrorHandler() (ident1 ErrorHandler) {
	invocation := new(ServiceInternalServerErrorHandlerInvocation)

	ident1 = _f7.InternalServerErrorHandlerHook()

	invocation.Results.Ident1 = ident1

	_f7.InternalServerErrorHandlerCalls = append(_f7.InternalServerErrorHandlerCalls, invocation)

	return
}

// InternalServerErrorHandlerCalled returns true if FakeService.InternalServerErrorHandler was called
func (f *FakeService) InternalServerErrorHandlerCalled() bool {
	return len(f.InternalServerErrorHandlerCalls) != 0
}

// AssertInternalServerErrorHandlerCalled calls t.Error if FakeService.InternalServerErrorHandler was not called
func (f *FakeService) AssertInternalServerErrorHandlerCalled(t *testing.T) {
	t.Helper()
	if len(f.InternalServerErrorHandlerCalls) == 0 {
		t.Error("FakeService.InternalServerErrorHandler not called, expected at least one")
	}
}

// InternalServerErrorHandlerNotCalled returns true if FakeService.InternalServerErrorHandler was not called
func (f *FakeService) InternalServerErrorHandlerNotCalled() bool {
	return len(f.InternalServerErrorHandlerCalls) == 0
}

// AssertInternalServerErrorHandlerNotCalled calls t.Error if FakeService.InternalServerErrorHandler was called
func (f *FakeService) AssertInternalServerErrorHandlerNotCalled(t *testing.T) {
	t.Helper()
	if len(f.InternalServerErrorHandlerCalls) != 0 {
		t.Error("FakeService.InternalServerErrorHandler called, expected none")
	}
}

// InternalServerErrorHandlerCalledOnce returns true if FakeService.InternalServerErrorHandler was called exactly once
func (f *FakeService) InternalServerErrorHandlerCalledOnce() bool {
	return len(f.InternalServerErrorHandlerCalls) == 1
}

// AssertInternalServerErrorHandlerCalledOnce calls t.Error if FakeService.InternalServerErrorHandler was not called exactly once
func (f *FakeService) AssertInternalServerErrorHandlerCalledOnce(t *testing.T) {
	t.Helper()
	if len(f.InternalServerErrorHandlerCalls) != 1 {
		t.Errorf("FakeService.InternalServerErrorHandler called %d times, expected 1", len(f.InternalServerErrorHandlerCalls))
	}
}

// InternalServerErrorHandlerCalledN returns true if FakeService.InternalServerErrorHandler was called at least n times
func (f *FakeService) InternalServerErrorHandlerCalledN(n int) bool {
	return len(f.InternalServerErrorHandlerCalls) >= n
}

// AssertInternalServerErrorHandlerCalledN calls t.Error if FakeService.InternalServerErrorHandler was called less than n times
func (f *FakeService) AssertInternalServerErrorHandlerCalledN(t *testing.T, n int) {
	t.Helper()
	if len(f.InternalServerErrorHandlerCalls) < n {
		t.Errorf("FakeService.InternalServerErrorHandler called %d times, expected >= %d", len(f.InternalServerErrorHandlerCalls), n)
	}
}
