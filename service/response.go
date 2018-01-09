package service

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/percolate/shisa/contenttype"
)

const (
	LocationHeaderKey = "Location"
)

var (
	jsonContentType = contenttype.ApplicationJson.String()
)

//go:generate charlatan -output=./response_charlatan.go Response

type Response interface {
	StatusCode() int
	Headers() http.Header
	Trailers() http.Header
	Serialize(io.Writer) (int, error)
}

type BasicResponse struct {
	Code     int
	headers  http.Header
	trailers http.Header
}

func (r *BasicResponse) StatusCode() int {
	return r.Code
}

func (r *BasicResponse) Headers() http.Header {
	if r.headers == nil {
		r.headers = make(http.Header)
	}
	return r.headers
}

func (r *BasicResponse) Trailers() http.Header {
	if r.trailers == nil {
		r.trailers = make(http.Header)
	}
	return r.trailers
}

func (r *BasicResponse) Serialize(io.Writer) (int, error) {
	return 0, nil
}

type JsonResponse struct {
	BasicResponse
	Payload json.Marshaler
}

func (r *JsonResponse) Serialize(w io.Writer) (int, error) {
	writer := countingWriter{delegate: w}
	encoder := json.NewEncoder(&writer)
	encoder.SetIndent("", "")
	encoder.SetEscapeHTML(true)

	err := encoder.Encode(r.Payload)
	return writer.count, err
}

func NewEmpty(code int) Response {
	return &BasicResponse{
		Code:     code,
		headers:  make(http.Header),
		trailers: make(http.Header),
	}
}

func NewOK(body json.Marshaler) Response {
	headers := make(http.Header)
	headers.Set(contenttype.ContentTypeHeaderKey, jsonContentType)

	return &JsonResponse{
		BasicResponse: BasicResponse{
			Code:     http.StatusOK,
			headers:  headers,
			trailers: make(http.Header),
		},
		Payload: body,
	}
}

type countingWriter struct {
	delegate io.Writer
	count    int
}

func (c *countingWriter) Write(p []byte) (n int, err error) {
	n, err = c.delegate.Write(p)
	c.count += n

	return
}

func NewSeeOther(location string) Response {
	headers := make(http.Header)
	headers.Set(LocationHeaderKey, location)
	return &BasicResponse{
		Code:     http.StatusSeeOther,
		headers:  headers,
		trailers: make(http.Header),
	}
}

func NewTemporaryRedirect(location string) Response {
	headers := make(http.Header)
	headers.Set(LocationHeaderKey, location)
	return &BasicResponse{
		Code:     http.StatusTemporaryRedirect,
		headers:  headers,
		trailers: make(http.Header),
	}
}
