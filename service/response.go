package service

import (
	"encoding/json"
	"io"
	"net/http"
)

//go:generate charlatan -output=./response_charlatan.go Response

type Response interface {
	StatusCode() int
	Headers() http.Header
	Trailers() http.Header
	Serialize(io.Writer) error
}

type basicResponse struct {
	status   int
	headers  http.Header
	trailers http.Header
}

func (r *basicResponse) StatusCode() int {
	return r.status
}

func (r *basicResponse) Headers() http.Header {
	return r.headers
}

func (r *basicResponse) Trailers() http.Header {
	return r.trailers
}

func (r *basicResponse) Serialize(io.Writer) error {
	return nil
}

type jsonResponse struct {
	basicResponse
	payload json.Marshaler
}

func (r *jsonResponse) Serialize(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "")
	encoder.SetEscapeHTML(true)

	return encoder.Encode(r.payload)
}

func NewEmpty(status int) Response {
	return &basicResponse{
		status:   status,
		headers:  make(http.Header),
		trailers: make(http.Header),
	}
}

func NewOK(body json.Marshaler) Response {
	return &jsonResponse{
		basicResponse: basicResponse{
			status:   http.StatusOK,
			headers:  make(http.Header),
			trailers: make(http.Header),
		},
		payload: body,
	}
}
