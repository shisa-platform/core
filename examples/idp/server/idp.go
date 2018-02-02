package idp

import (
	"errors"
	"expvar"
	"strings"

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
	Value string
}

type Idp struct {
	Logger *zap.Logger
	Hits   *expvar.Map
}

func (s *Idp) AuthenticateToken(message *Message, reply *string) (err error) {
	s.Hits.Add("AuthenticateToken", 1)
	defer func() {
		s.Logger.Info("AuthenticateToken", zap.String("request-id", message.RequestID), zap.Bool("OK", reply != nil), zap.Error(err))
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

func (s *Idp) FindUser(message *Message, reply *User) (err error) {
	s.Hits.Add("FindUser", 1)

	for _, user := range users {
		if user.Ident == message.Value {
			*reply = user
			break
		}
	}

	s.Logger.Info("FindUser", zap.String("request-id", message.RequestID), zap.String("user-id", message.Value), zap.Bool("found", reply != nil))
	return
}

func (s *Idp) Healthcheck(_ bool, reply *bool) (err error) {
	s.Hits.Add("Healthcheck", 1)
	*reply = true

	s.Logger.Info("Healthcheck", zap.Bool("ready", true))
	return
}
