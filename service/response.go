package service

import (
	"io"
	"net/http"
)

//go:generate charlatan -output=./response_charlatan.go Response

type Response interface {
	StatusCode() int
	Header() http.Header
	Trailer() http.Header
	Serialize(io.Writer) error
}
