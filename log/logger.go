package log

import (
	"fmt"
	"log"
	"log/syslog"
	"net/http"
	"os"
)

// xxx - LoggingProvider.writeLog(context, level, message), Syslog

type Logger interface {
	Info(requestID, message string)
	Infof(requestID, format string, args ...interface{})
	Error(requestID, message string)
	Errorf(requestID, format string, args ...interface{})
	Trace(requestID, message string)
	Tracef(requestID, format string, args ...interface{})
	// xxx - remove stdout, stderr here?
	Stdout() *log.Logger
	Stderr() *log.Logger
	Close()
}

type syslogLogger struct {
	trace  bool
	logger *syslog.Writer
	stdout *log.Logger
	stderr *log.Logger
}

// xxx - pass in logwriter factory?

func New(prefix string) (Logger, error) {
	logger, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER, prefix)
	if err != nil {
		return nil, fmt.Errorf("Unable to open syslog: %v", err)
	}

	return &syslogLogger{
		logger: logger,
		stdout: log.New(NewLogWriter(os.Stdout, prefix), "", 0),
		stderr: log.New(NewLogWriter(os.Stderr, prefix), "", 0),
	}, nil
}

func (l *syslogLogger) Info(requestID, message string) {
	l.writeLog(l.logger.Info, requestID, message)
}

func (l *syslogLogger) Infof(requestID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.writeLog(l.logger.Info, requestID, message)
}

func (l *syslogLogger) Error(requestID, message string) {
	l.writeLog(l.logger.Err, requestID, message)
}

func (l *syslogLogger) Errorf(requestID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.writeLog(l.logger.Err, requestID, message)
}

func (l *syslogLogger) Trace(requestID, message string) {
	if !l.trace {
		return
	}
	l.stderr.Print(requestID, " ", message)
}

func (l *syslogLogger) Tracef(requestID, format string, args ...interface{}) {
	if !l.trace {
		return
	}
	message := fmt.Sprintf(format, args...)
	l.stderr.Print(requestID, " ", message)
}

func (l *syslogLogger) Stdout() *log.Logger {
	return l.stdout
}

func (l *syslogLogger) Stderr() *log.Logger {
	return l.stderr
}

func (l *syslogLogger) Close() {
	l.logger.Close()
}

func (l *syslogLogger) writeLog(logger func(string) error, requestID, message string) {
	var msg string
	if requestID == "" {
		msg = "- " + message
	} else {
		msg = requestID + " " + message
	}

	if err := logger(msg); err != nil {
		l.stderr.Printf("ERROR! Unable to write to syslog: %v", err)
		l.stderr.Print(msg)
	}
}
