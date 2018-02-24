package httpx

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/percolate/shisa/contenttype"
)

const (
	LocationHeaderKey = "Location"
)

var (
	jsonContentType = contenttype.ApplicationJson.String()
	bufPool         = sync.Pool{
		New: func() interface{} {
			return make([]byte, 32*1024)
		},
	}
)

//go:generate charlatan -output=./response_charlatan.go Response

type Response interface {
	StatusCode() int
	Headers() http.Header
	Trailers() http.Header
	Err() error
	Serialize(io.Writer) error
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

func (r *BasicResponse) Err() error {
	return nil
}

func (r *BasicResponse) Serialize(io.Writer) error {
	return nil
}

type JsonResponse struct {
	BasicResponse
	Payload json.Marshaler
}

func (r *JsonResponse) Serialize(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "")
	encoder.SetEscapeHTML(true)

	return encoder.Encode(r.Payload)
}

func NewEmpty(code int) Response {
	return &BasicResponse{
		Code: code,
	}
}

type ErrorResponse struct {
	BasicResponse
	Error error
}

func (r *ErrorResponse) Err() error {
	return r.Error
}

func NewEmptyError(code int, err error) Response {
	return &ErrorResponse{
		BasicResponse: BasicResponse{
			Code: code,
		},
		Error: err,
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

// ResponseAdapter is an adapter for `http.Response` to the
// `Response` interface.
type ResponseAdapter struct {
	*http.Response
}

func (r ResponseAdapter) StatusCode() int {
	return r.Response.StatusCode
}

func (r ResponseAdapter) Headers() http.Header {
	return r.Header
}

func (r ResponseAdapter) Trailers() http.Header {
	return r.Trailer
}

func (r ResponseAdapter) Err() error {
	return nil
}

func (r ResponseAdapter) Serialize(w io.Writer) error {
	buf := getBuffer()
	defer putBuffer(buf)

	_, err := io.CopyBuffer(w, r.Body, buf)
	r.Body.Close()

	return err
}

func getBuffer() []byte {
	buf := bufPool.Get().([]byte)
	buf = buf[:cap(buf)]
	return buf
}

func putBuffer(buf []byte) {
	bufPool.Put(buf)
}
