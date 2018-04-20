package crash

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

var NoopReporter Reporter = noopReporter{}

// Reporter sends a crash report to an external service
type Reporter interface {
	Report(context.Context, *httpx.Request, merry.Error)
	Close() merry.Error
}

type noopReporter struct{}

func (r noopReporter) Report(context.Context, *httpx.Request, merry.Error) {
	return
}

func (r noopReporter) Close() merry.Error {
	return nil
}
