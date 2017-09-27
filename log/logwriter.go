package logx

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	logTimeFormat = "Jan 2 15:04:05"
)

type LogWriter struct {
	writer   io.Writer
	hostname string
	prefix   string
	pid      int
}

func NewLogWriter(writer io.Writer, prefix string) *LogWriter {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "-"
	}
	return &LogWriter{writer, hostname, prefix, os.Getpid()}
}

func (w *LogWriter) Write(bytes []byte) (int, error) {
	now := time.Now().UTC().Format(logTimeFormat)
	return fmt.Fprintf(w.writer, "%s %s %s[%d]: %s", now, w.hostname, w.prefix, w.pid, string(bytes))
}
