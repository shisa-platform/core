package gateway

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/requests"
	"github.com/percolate/shisa/responses"
)

func (s *Service) dispatch(w http.ResponseWriter, r *http.Request) {
	// xxx MetricsProvider? TracingProvider?

	rw := responseWrapper{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
	// xxx - require a parent context and delegate to that, here use timeout ctx
	c := &ctx.Context{
		Logger:  s.logger,
		Metrics: metrics.New(),
		Tracing: s.trace,
	}

	defer s.recovery(c, w, r, startTime, timer)

	var err error
	// do something
	rw.Write([]byte("hello, world"))

	rw.flush()
	go s.emitLogs(c, r, &rw, startTime, elapsed, err)
}

func (s *Service) recovery(c *ctx.Context, w http.ResponseWriter, r *http.Request, start time.Time, timer metrics.Timer) {
	cause := recover()
	if cause == nil {
		return
	}
	elapsed := float64(timer.Stop()) / 1e6

	msg := fmt.Errorf("panic: %v", cause)
	response := responses.NewInternalServerError(msg, responses.ErrorCodingException)

	written, err := response.Emit(w)
	if err != nil {
		format := "recovery emit failed: %v (original error: %v)"
		// xxx - use nested errors library?
		response.SetCause(fmt.Errorf(format, err, response.GetCause()))
	}

	snap := buildSnapshot(c, r, response.Status, written, start, elapsed)
	s.emitSnapshot(snap, response)

	// dump stack trace to stderr
	s.logger.Stderr().Printf("%s panic: %v\n", c.ID, cause)
	debug.PrintStack()
}

func (s *Service) emitLogs(c *ctx.Context, r *http.Request, rw *responseWrapper, start time.Time, duration float64, err error) {
	snap := buildSnapshot(c, r, rw.status, rw.bytesWritten, start, duration)
	s.emitSnapshot(snap, err)
}

func (s *Service) emitSnapshot(snap *Snapshot, err error) {
	s.emitRequestLog(snap)
	s.emitRequestError(snap, err)
	s.emitRequestReferrer(snap)
	s.emitRequestLatency(snap)
	s.emitStats(snap)
}

func (s *Service) emitRequestLog(snap *Snapshot) {
	s.logger.Info(snap.requestID, snap.String())
}

func (s *Service) emitRequestError(snap *Snapshot, err error) {
	if _, ok := err.(responses.ErrorResponder); ok {
		s.logger.Error(snap.requestID, err.Error())
	}
	if snap.statusCode >= 500 && snap.statusCode <= 599 {
		// send to exception capture provider
	}
}

func (s *Service) emitRequestReferrer(snap *Snapshot) {
	referrer := snap.request.Header.Get("Referer")
	if referrer == "" {
		return
	}
	s.logger.Infof(snap.requestID, "referrer: %s", referrer)
}

func (s *Service) emitUpstreamLatency(snap *Snapshot) {
	// xxx - upstream start provider
	if start, err := upstreamStartProvider(snap.request); err == nil {
		// this latency calculation is only a rough approximation
		latency := int64((snap.startTime.UnixNano() / 1e6) - start)
		s.logger.Infof(
			snap.requestID,
			"upstream start: %s, latency ~= %d",
			startHeader,
			latency)
	}
}

func (s *Service) emitStats(snap *Snapshot) {
	// xxx - publish to tracing provider? is that automated?
	// xxx - forward to additional metrics capture provider(s)?
}

type responseWrapper struct {
	http.ResponseWriter
	status       int
	bytesWritten int
	buffer       bytes.Buffer
}

func (w *responseWrapper) Write(data []byte) (int, error) {
	written, err := w.buffer.Write(data)
	w.bytesWritten += written

	return written, err
}

func (w *responseWrapper) WriteHeader(status int) {
	w.status = status
}

func (w *responseWrapper) reset() {
	w.buffer.Reset()
	w.bytesWritten = 0
	w.status = http.StatusOK
}

func (w *responseWrapper) flush() error {
	w.ResponseWriter.WriteHeader(w.status)
	written, err := w.ResponseWriter.Write(w.buffer.Bytes())
	w.bytesWritten = written

	return err
}
