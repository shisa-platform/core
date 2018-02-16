package log

import (
	"time"

	"go.uber.org/zap/zapcore"
)

const (
	ActorIDKey      = "actor-id"
	ClientIPKey     = "client-ip-address"
	ElapsedTimeKey  = "elapsed"
	MethodKey       = "method"
	RequestIDKey    = "request-id"
	ResponseSizeKey = "response-size"
	StartTimeKey    = "start"
	StatusCodeKey   = "status-code"
	ServiceNameKey  = "service"
	URIKey          = "uri"
	UserAgentKey    = "user-agent"
)

func ActorID(val string) zapcore.Field {
	return zapcore.Field{Key: ActorIDKey, Type: zapcore.StringType, String: val}
}

func ClientIP(val string) zapcore.Field {
	return zapcore.Field{Key: ClientIPKey, Type: zapcore.StringType, String: val}
}

func ElapsedTime(val time.Duration) zapcore.Field {
	return zapcore.Field{Key: ElapsedTimeKey, Type: zapcore.DurationType, Integer: int64(val)}
}

func Method(val string) zapcore.Field {
	return zapcore.Field{Key: MethodKey, Type: zapcore.StringType, String: val}
}

func RequestID(val string) zapcore.Field {
	return zapcore.Field{Key: RequestIDKey, Type: zapcore.StringType, String: val}
}

func ResponseSize(val int) zapcore.Field {
	return zapcore.Field{Key: ResponseSizeKey, Type: zapcore.Int64Type, Integer: int64(val)}
}

func ServiceName(val string) zapcore.Field {
	return zapcore.Field{Key: ServiceNameKey, Type: zapcore.StringType, String: val}
}

func StartTime(val time.Time) zapcore.Field {
	return zapcore.Field{Key: StartTimeKey, Type: zapcore.TimeType, Integer: val.UnixNano(), Interface: val.Location()}
}

func StatusCode(val int) zapcore.Field {
	return zapcore.Field{Key: StatusCodeKey, Type: zapcore.Int64Type, Integer: int64(val)}
}

func URI(val string) zapcore.Field {
	return zapcore.Field{Key: URIKey, Type: zapcore.StringType, String: val}
}

func UserAgent(val string) zapcore.Field {
	return zapcore.Field{Key: UserAgentKey, Type: zapcore.StringType, String: val}
}
