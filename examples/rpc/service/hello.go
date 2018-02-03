package hello

import (
	"errors"
	"fmt"
	"net/rpc"

	"go.uber.org/zap"

	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/examples/idp/service"
)

const idpServiceAddrEnv = "IDP_SERVICE_ADDR"

type Message struct {
	RequestID string
	UserID    string
	Language  string
	Name      string
}

type Hello struct {
	Logger *zap.Logger
}

func (s *Hello) Greeting(message *Message, reply *string) (err error) {
	defer func() {
		r := ""
		if reply != nil {
			r = *reply
		}
		s.Logger.Info("Greeting", zap.String("request-id", message.RequestID), zap.String("language", message.Language), zap.String("user-id", message.UserID), zap.String("reply", r), zap.Error(err))
	}()

	client, err := connect()
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

	client, err := connect()
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

func connect() (*rpc.Client, error) {
	addr, envErr := env.Get(idpServiceAddrEnv)
	if envErr != nil {
		return nil, envErr
	}

	client, rpcErr := rpc.DialHTTP("tcp", addr)
	if rpcErr != nil {
		return nil, rpcErr
	}

	return client, nil
}
