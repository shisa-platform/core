package service

import (
	"net/http"
)

func (s *Service) provision() http.Handler {
	return http.HandlerFunc(s.dispatch)
}
