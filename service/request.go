package service

import (
	"crypto/rand"
	"net/http"
	"strings"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/uuid"
)

// StringExtractor is a function type that extracts a string from
// the given `context.Context` and `*service.Request`.
// An error is returned if the string could not be extracted.
type StringExtractor func(context.Context, *Request) (string, merry.Error)

type Request struct {
	*http.Request
	PathParams  []PathParameter
	QueryParams []QueryParameter
	id          string
	clientIP    string
}

func (r *Request) PathParamExists(name string) bool {
	for _, p := range r.PathParams {
		if p.Name == name {
			return true
		}
	}

	return false
}

func (r *Request) QueryParamExists(name string) bool {
	for _, p := range r.QueryParams {
		if p.Name == name {
			return true
		}
	}

	return false
}

// ID returns a globally unique string for the request.
// It creates a version 5 UUID with the concatenation of current
// unix nanos, three bytes of random data, the client ip address,
// the request method and the request URI.
// This method is idempotent.
func (r *Request) ID() string {
	if r.id == "" {
		r.id = GenerateID(r.Request)
	}

	return r.id
}

// ClientIP attempts to extract the IP address of the user agent
// from the request.
// The "X-Real-IP" and "X-Forwarded-For" headers are checked
// followed by the `RemoteAddr` field of the request.  An empty
// string will be returned if nothing can be found.
func (r *Request) ClientIP() string {
	if r.clientIP == "" {
		r.clientIP = ClientIP(r.Request)
	}

	return r.clientIP
}

// GenerateID creates a globally unique string for the request.
// It creates a version 5 UUID with the concatenation of current
// unix nanos, three bytes of random data, the client ip address,
// the request method and the request URI.
// This function is *not* idempotent.
func GenerateID(r *http.Request) string {
	now := time.Now().UnixNano()
	clientAddr := ClientIP(r)

	// The following logic is roughly equivilent to:
	// `fmt.Sprintf("%v%x%v%v%v", now, nonce, clientAddr, r.Method, r.RequestURI)`
	// N.B. - sizeof(int64) + 3 (nonce length) = 11
	b := make([]byte, 11+len(clientAddr)+len(r.Method)+len(r.RequestURI))
	// N.B. - `now` is a `int64` so we can simply add those bytes to our array
	b[0] = byte(now)
	b[1] = byte(now >> 8)
	b[2] = byte(now >> 16)
	b[3] = byte(now >> 24)
	b[4] = byte(now >> 32)
	b[5] = byte(now >> 40)
	b[6] = byte(now >> 48)
	b[7] = byte(now >> 56)
	rand.Read(b[8:10])
	copy(b[11:], []byte(clientAddr))
	copy(b[11+len(clientAddr):], []byte(r.Method))
	copy(b[11+len(clientAddr)+len(r.Method):], []byte(r.RequestURI))

	return uuid.New(uuid.ShisaNS, string(b)).String()
}

// ClientIP attempts to extract the IP address of the user agent
// from the request.
// The "X-Real-IP" and "X-Forwarded-For" headers are checked
// followed by the `RemoteAddr` field of the request.  An empty
// string will be returned if nothing can be found.
func ClientIP(r *http.Request) string {
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
