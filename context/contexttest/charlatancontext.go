// Code generated by "charlatan -interfaces=Context -output=./contexttest/charlatancontext.go /usr/local/go/src/context/context.go"; DO NOT EDIT.

package contexttest

import "time"

////////////////////////////////
// This is a mock for Context //
////////////////////////////////
type DeadlineInvocation struct {
	Results struct {
		Deadline time.Time
		Ok       bool
	}
}

type DoneInvocation struct {
	Results struct {
		Ret0 <-chan struct{}
	}
}

type ErrInvocation struct {
	Results struct {
		Ret0 error
	}
}

type ValueInvocation struct {
	Parameters struct {
		Key interface{}
	}
	Results struct {
		Ret0 interface{}
	}
}

type FakeContext struct {
	DeadlineHook func() (deadline time.Time, ok bool)
	DoneHook     func() (ret0 <-chan struct{})
	ErrHook      func() (ret0 error)
	ValueHook    func(key interface{}) (ret0 interface{})

	DeadlineCalls []*DeadlineInvocation
	DoneCalls     []*DoneInvocation
	ErrCalls      []*ErrInvocation
	ValueCalls    []*ValueInvocation
}

func (a *FakeContext) Deadline() (deadline time.Time, ok bool) {
	invocation := new(DeadlineInvocation)

	deadline, ok = a.DeadlineHook()

	invocation.Results.Deadline = deadline
	invocation.Results.Ok = ok

	a.DeadlineCalls = append(a.DeadlineCalls, invocation)

	return deadline, ok
}

func (a *FakeContext) Done() (ret0 <-chan struct{}) {
	invocation := new(DoneInvocation)

	ret0 = a.DoneHook()

	invocation.Results.Ret0 = ret0

	a.DoneCalls = append(a.DoneCalls, invocation)

	return ret0
}

func (a *FakeContext) Err() (ret0 error) {
	invocation := new(ErrInvocation)

	ret0 = a.ErrHook()

	invocation.Results.Ret0 = ret0

	a.ErrCalls = append(a.ErrCalls, invocation)

	return ret0
}

func (a *FakeContext) Value(key interface{}) (ret0 interface{}) {
	invocation := new(ValueInvocation)

	invocation.Parameters.Key = key

	ret0 = a.ValueHook(key)

	invocation.Results.Ret0 = ret0

	a.ValueCalls = append(a.ValueCalls, invocation)

	return ret0
}
