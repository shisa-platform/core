package httpx

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

// ResponseInterceptor implements `http.ResponseWriter` to capture
// and log the response sent to a user agent.
// Use it to wrap a standard  `http.ResponseWriter` and log the
// request to a logger named "request" at the `Info` level.
type ResponseInterceptor struct {
	http.ResponseWriter
	logger   *zap.Logger
	start    time.Time
	status   int
	size     int
}

func (i *ResponseInterceptor) Header() http.Header {
	return i.ResponseWriter.Header()
}

func (i *ResponseInterceptor) Write(data []byte) (int, error) {
	size, err := i.ResponseWriter.Write(data)
	i.size += size

	return size, err
}

func (i *ResponseInterceptor) WriteHeader(status int) {
	i.status = status
	i.ResponseWriter.WriteHeader(status)
}

// Flush logs the request at the `Info` level.
// No logging is peformed the `Info` level is not configured.  If
// the underlying writer implements `http.Flusher` then the
// `Flush` method will be called.
func (i *ResponseInterceptor) Flush(ctx context.Context, request *service.Request) {
	if ce := i.logger.Check(zap.InfoLevel, "request"); ce != nil {
		end := time.Now().UTC()
		elapsed := end.Sub(i.start)
		if i.status == 0 {
			i.status = http.StatusOK
		}
		fs := make([]zapcore.Field, 9, 10)
		fs[0] = zap.String("request-id", ctx.RequestID())
		fs[1] = zap.String("client-ip-address", request.ClientIP())
		fs[2] = zap.String("method", request.Method)
		fs[3] = zap.String("uri", request.URL.RequestURI())
		fs[4] = zap.Int("status-code", i.status)
		fs[5] = zap.Int("response-size", i.size)
		fs[6] = zap.String("user-agent", request.UserAgent())
		fs[7] = zap.Time("start", i.start)
		fs[8] = zap.Duration("elapsed", elapsed)
		if u := ctx.Actor(); u != nil {
			fs = append(fs, zap.String("user-id", u.ID()))
		}
		ce.Write(fs...)
	}

	if f, ok := i.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func NewInterceptor(w http.ResponseWriter, logger *zap.Logger) *ResponseInterceptor {
	return &ResponseInterceptor{
		ResponseWriter: w,
		logger: logger,
		start: time.Now().UTC(),
	}
}
