package hello

import (
	"errors"
	"fmt"
	"net/rpc"

	"go.uber.org/zap"

	"github.com/percolate/shisa/examples/idp/service"
	"github.com/percolate/shisa/sd"
)

const idpServiceName = "idp"

type Message struct {
	RequestID string
	UserID    string
	Language  string
	Name      string
}

type Hello struct {
	Resolver sd.Resolver
	Logger   *zap.Logger
}

func (s *Hello) Greeting(message *Message, reply *string) (err error) {
	defer func() {
		r := ""
		if reply != nil {
			r = *reply
		}
		s.Logger.Info("Greeting", zap.String("request-id", message.RequestID), zap.String("language", message.Language), zap.String("user-id", message.UserID), zap.String("reply", r), zap.Error(err))
	}()

	client, err := s.connect()
	if err != nil {
		return
	}

	who := message.Name

	if who == "" {
		request := idp.Message{RequestID: message.RequestID, Value: message.UserID}
		var user idp.User
		err = client.Call("Idp.FindUser", &request, &user)
		if err != nil {
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
	defer func() {
		s.Logger.Info("Healthcheck", zap.String("request-id", requestID), zap.Bool("ready", *reply), zap.Error(err))
	}()

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
	nodes, resErr := s.Resolver.Resolve(idpServiceName)
	if resErr != nil {
		return nil, resErr
	}

	client, rpcErr := rpc.DialHTTP("tcp", nodes[0])
	if rpcErr != nil {
		return nil, rpcErr
	}

	return client, nil
}
