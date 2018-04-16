package crash

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

// Reporter sends a crash report to an external service
type Reporter interface {
	Report(context.Context, *httpx.Request, merry.Error)
	Close() merry.Error
}
