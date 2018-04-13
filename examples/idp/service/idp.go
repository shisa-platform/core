package idp

import (
	"errors"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
)

var (
	users = []User{
		User{"user:1", "Admin", "password"},
		User{"user:2", "Boss", "password"},
	}
)

type Message struct {
	RequestID string
	Value     string
	Metadata  map[string]string
}

type Idp struct {
	Logger *zap.Logger
}

func (s *Idp) AuthenticateToken(message *Message, reply *string) (err error) {
	span := s.startSpan(message, "AuthenticateToken")
	defer span.Finish()

	defer func() {
		s.Logger.Info("AuthenticateToken", zap.String("request-id", message.RequestID), zap.Bool("OK", reply != nil))
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.String("error", err.Error()))
			s.Logger.Error("AuthenticateToken", zap.String("request-id", message.RequestID), zap.String("error", err.Error()))
		}
	}()

	credentialParts := strings.Split(message.Value, ":")
	if len(credentialParts) != 2 {
		err = errors.New("wrong credential parts count")
		return
	}

	for _, user := range users {
		if user.Name == credentialParts[0] && user.Pass == credentialParts[1] {
			*reply = user.Ident
			break
		}
	}

	return
}

func (s *Idp) FindUser(message *Message, reply *User) error {
	span := s.startSpan(message, "FindUser")
	defer span.Finish()

	for _, user := range users {
		if user.Ident == message.Value {
			*reply = user
			break
		}
	}

	s.Logger.Info("FindUser", zap.String("request-id", message.RequestID), zap.String("user-id", message.Value), zap.Bool("found", reply != nil))
	return nil
}

func (s *Idp) Healthcheck(requestID string, reply *bool) error {
	*reply = true

	return nil
}

func (s *Idp) startSpan(message *Message, operation string) opentracing.Span {
	var span opentracing.Span
	carrier := opentracing.TextMapCarrier(message.Metadata)
	if spanContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier); err == nil {
		span = opentracing.StartSpan(operation, ext.RPCServerOption(spanContext))
	} else {
		tracer := opentracing.NoopTracer{}
		span = tracer.StartSpan(operation)
	}
	span.SetTag("request_id", message.RequestID)

	return span
}
