package hello

import (
	"fmt"

	"go.uber.org/zap"
)

type Message struct {
	RequestID string
	Language string
	Name     string
}

type Hello struct {
	Logger *zap.Logger
}

func (s *Hello) Greeting(message *Message, reply *string) error {
	greeting := greetings[AmericanEnglish]
	if msg, ok := greetings[message.Language]; ok {
		greeting = msg
	}

	*reply = fmt.Sprintf("%s %s", greeting, message.Name)

	s.Logger.Info("Greeting", zap.String("request-id", message.RequestID), zap.String("language", message.Language), zap.String("name", message.Name), zap.String("reply", *reply))

	return nil
}

func (s *Hello) Healthcheck(requestID string, reply *bool) error {
	*reply = true

	s.Logger.Info("Healthcheck", zap.String("request-id", requestID), zap.Bool("ready", true))
	return nil
}
