package service

import (
	"encoding/json"
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

type jsonResponse struct {
	status   int
	headers  http.Header
	trailers http.Header
	payload  json.Marshaler
}

func (r *jsonResponse) StatusCode() int {
	return r.status
}

func (r *jsonResponse) Header() http.Header {
	return r.headers
}

func (r *jsonResponse) Trailer() http.Header {
	return r.trailers
}

func (r *jsonResponse) Serialize(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "")
	encoder.SetEscapeHTML(true)

	return encoder.Encode(r.payload)
}

func NewOK(body json.Marshaler) Response {
	return &jsonResponse{
		status:   http.StatusOK,
		headers:  make(http.Header),
		trailers: make(http.Header),
		payload:  body,
	}
}
