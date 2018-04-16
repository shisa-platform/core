package crash

import (
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/getsentry/raven-go"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/errorx"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/models"
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
	client := &Fakecapturer{
		CaptureHook: func(*raven.Packet, map[string]string) (s string, ech chan error) {
			return
		},
	}
	sr := &SentryReporter{client}
	ctx := context.New(nil)
	r := httpx.GetRequest(httptest.NewRequest("GET", "/", nil))

	sr.Report(ctx, r, merry.New("not a panic"))

	client.AssertCaptureNotCalled(t)
}

func TestSentryReportPanic(t *testing.T) {
	client := &Fakecapturer{
		CaptureHook: func(*raven.Packet, map[string]string) (s string, ech chan error) {
			return
		},
	}
	sr := &SentryReporter{client}
	u := &models.FakeUser{
		IDHook: func() string {
			return "user_id"
		},
	}
	ctx := context.WithActor(nil, u).WithRequestID("req_id")
	r := httpx.GetRequest(httptest.NewRequest("GET", "/", nil))

	sr.Report(ctx, r, errorx.CapturePanic(merry.New("i blewed up"), "caused panic"))

	client.AssertCaptureCalledOnce(t)
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
