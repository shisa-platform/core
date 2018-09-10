package hello

import (
	"errors"
	"fmt"
	"net/rpc"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"

	"github.com/shisa-platform/core/examples/idp/service"
	"github.com/shisa-platform/core/lb"
)

const idpServiceName = "idp"

type Message struct {
	RequestID string
	UserID    string
	Language  string
	Name      string
	Metadata  map[string]string
}

type Hello struct {
	Balancer lb.Balancer
	Logger   *zap.Logger
}

func (s *Hello) Greeting(message *Message, reply *string) (err error) {
	span := s.startSpan(message)
	defer span.Finish()

	defer func() {
		r := ""
		if reply != nil {
			r = *reply
		}
		s.Logger.Info("Greeting", zap.String("request-id", message.RequestID), zap.String("user-id", message.UserID), zap.String("language", message.Language), zap.String("reply", r))
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.String("error", err.Error()))
			s.Logger.Error("Greeting", zap.String("request-id", message.RequestID), zap.String("error", err.Error()))
		}
	}()

	client, err := s.connect()
	if err != nil {
		return
	}

	who := message.Name

	if who == "" {
		request := idp.Message{
			RequestID: message.RequestID,
			Value:     message.UserID,
			Metadata:  make(map[string]string),
		}

		carrier := opentracing.TextMapCarrier(request.Metadata)
		if err = span.Tracer().Inject(span.Context(), opentracing.TextMap, carrier); err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.String("error", err.Error()))
			return
		}

		var user idp.User

		if err = client.Call("Idp.FindUser", &request, &user); err != nil {
			return
		}
		if user.Ident == "" {
			return errors.New("user not found")
		}
		who = user.Name
	}

	greeting := greetings[AmericanEnglish]
	if msg, ok := greetings[message.Language]; ok {
		greeting = msg
	}

	*reply = fmt.Sprintf("%s %s", greeting, who)

	return
}

func (s *Hello) Healthcheck(requestID string, reply *bool) (err error) {
	*reply = false

	client, err := s.connect()
	if err != nil {
		return
	}

	var idpReady bool
	arg := requestID
	err = client.Call("Idp.Healthcheck", &arg, &idpReady)
	if err != nil || !idpReady {
		return
	}

	*reply = true

	return
}

func (s *Hello) connect() (*rpc.Client, error) {
	node, resErr := s.Balancer.Balance(idpServiceName)
	if resErr != nil {
		return nil, resErr
	}

	client, rpcErr := rpc.DialHTTP("tcp", node)
	if rpcErr != nil {
		return nil, rpcErr
	}

	return client, nil
}

func (s *Hello) startSpan(message *Message) opentracing.Span {
	var span opentracing.Span
	carrier := opentracing.TextMapCarrier(message.Metadata)
	if spanContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier); err == nil {
		span = opentracing.StartSpan("Greeting", ext.RPCServerOption(spanContext))
	} else {
		s.Logger.Error("error extracting client trace", zap.String("request-id", message.RequestID), zap.String("error", err.Error()))
		span = opentracing.StartSpan("Greeting", opentracing.Tag{string(ext.SpanKind), "server"})
	}
	span.SetTag("request_id", message.RequestID)

	return span
}
