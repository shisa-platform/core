package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/percolate/shisa/uuid"
)

var (
	ParameterNotPresented = errors.New("parameter not presented")
)

type Request struct {
	*http.Request
	PathArgs    url.Values
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

func (r *Request) PathArg(name string) (string, bool) {
	if values, ok := r.PathArgs[name]; ok {
		return values[0], true
	}

	return "", false
}

func (r *Request) PathArgInt(name string) (int, error) {
	if values, ok := r.PathArgs[name]; ok {
		return strconv.Atoi(values[0])
	}

	return 0, ParameterNotPresented
}

func (r *Request) GenerateID() string {
	now := time.Now().UnixNano()
	nonce := make([]byte, 3)
	rand.Read(nonce)
	clientAddr := GetClientIP(r.Request)
	name := fmt.Sprintf("%v%x%v%v%v", now, nonce, clientAddr, r.Method, r.RequestURI)

	return uuid.New(uuid.ShisaNS, name).String()
}
