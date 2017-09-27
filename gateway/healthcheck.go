package service

import (
	"encoding/json"
	"net/http"

	"authn"
)

func (s *Service) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	connStats.Add("healthchecks", 1)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response, problem := ping(s.healthchecks)
	if problem {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if bytes, err := json.Marshal(response); err == nil {
		w.Write(bytes)
	}
}

func ping(pingables map[string]func() error) (map[string]string, bool) {
	result := make(map[string]string)
	var problem bool
	for name, pinger := range pingables {
		if err := pinger(); err != nil {
			problem = true
			result[name] = err.Error()
		} else {
			result[name] = "OK"
		}
	}

	return result, problem
}
