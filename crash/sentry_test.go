package crash

import (
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/models"
)

func TestNewSentryReporter(t *testing.T) {
	sr, err := NewSentryReporter("https://key:secret@localhost/project")

	assert.NoError(t, err)
	assert.NotNil(t, sr)
	assert.IsType(t, &SentryReporter{}, sr)
}

func TestNewSentryReporterError(t *testing.T) {
	sr, err := NewSentryReporter("not a dsn")

	assert.Error(t, err)
	assert.Nil(t, sr)
}

func TestSentryReport(t *testing.T) {
	client := &Fakecapturer{}
	client.SetCaptureStub("", nil)

	sr := &SentryReporter{client}
	ctx := context.New(nil)
	r := httpx.GetRequest(httptest.NewRequest("GET", "/", nil))

	sr.Report(ctx, r, merry.New("not a panic"))

	client.AssertCaptureNotCalled(t)
}

func TestSentryReportPanic(t *testing.T) {
	client := &Fakecapturer{}
	client.SetCaptureStub("", nil)

	sr := &SentryReporter{client}

	u := &models.FakeUser{}
	u.SetIDStub("user_id")
	ctx := context.WithActor(nil, u).WithRequestID("req_id")
	r := httpx.GetRequest(httptest.NewRequest("GET", "/", nil))

	var exception merry.Error

	defer func() {
		sr.Report(ctx, r, exception)
		client.AssertCaptureCalledOnce(t)
	}()

	defer errorx.CapturePanic(&exception, "uh-oh")

	panic(merry.New("i blewed up!"))
}

func TestClose(t *testing.T) {
	client := &Fakecapturer{
		CloseHook: func() { return },
	}
	sr := &SentryReporter{client}

	err := sr.Close()

	assert.NoError(t, err)
	client.AssertCloseCalledOnce(t)
}

func TestClosePanic(t *testing.T) {
	client := &Fakecapturer{
		CloseHook: func() { panic("i blewed up") },
	}
	sr := &SentryReporter{client}

	err := sr.Close()

	assert.Error(t, err)
	assert.True(t, errorx.IsPanic(err))
	client.AssertCloseCalledOnce(t)
}
