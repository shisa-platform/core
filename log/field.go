package log

import (
	"go.uber.org/zap/zapcore"
)

const (
	RequestIDKey = "request-id"
	ClientIPKey = "client-ip-address"
	MethodKey = "method"
	URIKey = "uri"
	StatusCodeKey = "status-code"
	ResponseSizeKey = "response-size"
	UserAgentKey = "user-agent"
	StartTimeKey = "start"
	ElapsedTimeKey = "elapsed"
	ServiceNameKey = "service"
	ActorIDKey = "actor-id"
)

func RequestID(val string) zapcore.Field {
	return zapcore.Field{Key: RequestIDKey, Type: zapcore.StringType, String: val}
}
