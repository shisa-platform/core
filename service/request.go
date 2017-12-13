package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/percolate/shisa/uuid"
)

var (
	ParameterNotPresented = errors.New("parameter not presented")
)

// Param is a single URL parameter, consisting of a key and a
// value.
type Param struct {
	Key   string
	Value string
}

// Params is a ordered slice of URL parameters.
type Params []Param

type Request struct {
	*http.Request
	PathParams  Params
	QueryParams url.Values
}

func (r *Request) QueryParamExists(name string) bool {
	_, ok := r.QueryParams[name]
	return ok
}

func (r *Request) QueryParam(name string) (string, bool) {
	if values, ok := r.QueryParams[name]; ok {
		return values[0], true
	}

	return "", false
}

func (r *Request) QueryParamBool(name string) (bool, bool) {
	if values, ok := r.QueryParams[name]; ok {
		b, err := strconv.ParseBool(values[0])
		if err != nil {
			return false, true
		}

		return b, true
	}

	return false, false
}

func (r *Request) QueryParamInt(name string) (int, error) {
	if values, ok := r.QueryParams[name]; ok {
		return strconv.Atoi(values[0])
	}

	return 0, ParameterNotPresented
}

func (r *Request) QueryParamUint(name string) (uint, error) {
	if values, ok := r.QueryParams[name]; ok {
		u64, err := strconv.ParseUint(values[0], 10, 0)
		return uint(u64), err
	}

	return 0, ParameterNotPresented
}

func (r *Request) PathParam(name string) (string, bool) {
	for _, p := range r.PathParams {
		if p.Key == name {
			return p.Value, true
		}
	}

	return "", false
}

func (r *Request) PathParamInt(name string) (int, error) {
	if v, ok := r.PathParam(name); ok {
		return strconv.Atoi(v)
	}

	return 0, ParameterNotPresented
}

// GenerateID creates a globally unique string for the request.
// It creates a version 5 UUID with the concatenation of current unix nanos,
// three bytes of random data, the client ip address, the request method and the
// request URI.
func (r *Request) GenerateID() string {
	now := time.Now().UnixNano()
	nonce := make([]byte, 3)
	rand.Read(nonce)
	clientAddr := r.ClientIP()
	name := fmt.Sprintf("%v%x%v%v%v", now, nonce, clientAddr, r.Method, r.RequestURI)

	return uuid.New(uuid.ShisaNS, name).String()
}

func (r *Request) ClientIP() string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		if ip = r.Header.Get("X-Forwarded-For"); ip == "" {
			ip = r.RemoteAddr
		}
	}
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}
