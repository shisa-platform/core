package stats

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"time"
)

const (
	format = "endpoint.%s.%s.status.%s:1|c\nendpoint.%[1]s.%[2]s.time:%.0[4]f|ms"
)

type StatsdService struct {
	service  string
	location *url.URL
}

func OpenStatsdService(location, service string) (Service, error) {
	uri, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	if uri.Scheme != "statsd" {
		return nil, fmt.Errorf("Invalid statsd URL: %q", location)
	}

	_, _, err = net.SplitHostPort(uri.Host)
	if err != nil {
		return nil, err
	}

	return &StatsdService{service: service, location: uri}, nil
}

func (s *StatsdService) PublishMetric(name string, code int, duration float64) error {
	status := "ok"
	if code >= 500 && code <= 599 {
		status = "er"
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, format, s.service, name, status, duration)
	w, err := net.DialTimeout("udp", s.location.Host, 5*time.Second)
	if err != nil {
		return fmt.Errorf("unable to send statsd packet: %s", err.Error())
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("unable to send statsd packet: %s", err.Error())
	}
	w.Close()

	return nil
}
