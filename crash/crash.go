package crash

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

// Reporter reports an error to an external service
type Reporter interface {
	Report(context.Context, *httpx.Request, merry.Error)
	Close() merry.Error
}
