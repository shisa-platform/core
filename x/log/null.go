package logx

import "log"

type null struct{}

var (
	NULL       = new(null)
	nullLogger = log.New(NULL, "", 0)
)

func (l *null) Info(string, string)                   {}
func (l *null) Infof(string, string, ...interface{})  {}
func (l *null) Error(string, string)                  {}
func (l *null) Errorf(string, string, ...interface{}) {}
func (l *null) Trace(string, string)                  {}
func (l *null) Tracef(string, string, ...interface{}) {}

func (l *null) SentryError(data SentryMetadata, err error) {}

func (l *null) Stdout() *log.Logger {
	return nullLogger
}

func (l *null) Stderr() *log.Logger {
	return nullLogger
}

func (l *null) Close() {}

func (l *null) Write(p []byte) (n int, err error) {
	return len(p), nil
}
