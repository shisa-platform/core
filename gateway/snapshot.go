package service

import (
	"fmt"
	"net/http"
	"time"

	"ctx"
)

type Snapshot struct {
	requestID    string
	ipAddress    string
	userID       string
	statusCode   int
	responseSize int
	startTime    time.Time
	elapsedTime  float64
	request      *http.Request
}

func (r *Snapshot) String() string {
	return fmt.Sprintf(
		// request-id ip-addr tenant-id user-id timestamp "METHOD uri" status bytes time-ms "user-agent"
		`%s %s %d "%s %s" %d %d %.3f "%s"`,
		r.ipAddress,
		r.userID,
		r.startTime.UnixNano()/1e6,
		r.request.Method,
		r.request.URL.RequestURI(),
		r.statusCode,
		r.responseSize,
		r.elapsedTime,
		r.request.UserAgent(),
	)
}

func (r *Snapshot) GetUser() string {
	return r.userID
}

func (r *Snapshot) GetRequest() *http.Request {
	return r.request
}

func (r *Snapshot) GetTime() float64 {
	return r.elapsedTime
}

func (r *Snapshot) GetRequestID() string {
	return r.requestID
}

func buildSnapshot(c *ctx.Context, r *http.Request, status, size int, start time.Time, duration float64) *Snapshot {
	return &Snapshot{
		requestID:    c.ID,
		ipAddress:    getClientAddressOrDefault(r),
		userID:       getUserIDOrDefault(c),
		statusCode:   status,
		responseSize: size,
		startTime:    start,
		elapsedTime:  duration,
		metrics:      c.Metrics,
		request:      r,
		// xxx - figure out how to pass the metric name around?
		//metricName:   ctx.GetHandlerMetricName(c),
	}
}

func GetClientIP(req *http.Request) string {
	ip := req.Header.Get("X-Real-IP")
	if ip == "" {
		if ip = req.Header.Get("X-Forwarded-For"); ip == "" {
			ip = req.RemoteAddr
		}
	}
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	return ip
}

func getClientAddressOrDefault(r *http.Request) string {
	clientAddress := GetClientIP(r)
	if clientAddress == "" {
		clientAddress = "-"
	}

	return clientAddress
}

func getUserIDOrDefault(c *ctx.Context) string {
	userID := "-"
	if c.Actor != nil {
		userID = c.Actor.String()
	}

	return userID
}
