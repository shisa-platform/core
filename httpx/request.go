package httpx

import (
	"crypto/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/uuid"
)

var (
	requestPool = sync.Pool{
		New: func() interface{} {
			return new(Request)
		},
	}
	InvalidParameterNameEscape  = merry.New("invalid parameter name escape")
	InvalidParameterValueEscape = merry.New("invalid parameter value escape")
	MalformedQueryParamter      = merry.New("malformed query parameter")
	MissingQueryParamter        = merry.New("missing query parameter")
	UnknownQueryParamter        = merry.New("unknown query parameter")
)

// GetRequest returns a Request instance from the shared pool,
// ready for (re)use.
func GetRequest(parent *http.Request) *Request {
	request := requestPool.Get().(*Request)
	request.Request = parent
	request.PathParams = nil
	request.QueryParams = nil
	request.id = ""
	request.clientIP = ""

	return request
}

// PutRequest returns the given Request back to the shared pool.
func PutRequest(request *Request) {
	requestPool.Put(request)
}

type Request struct {
	*http.Request
	PathParams  []PathParameter
	QueryParams []*QueryParameter
	id          string
	clientIP    string
}

// ParseQueryParameters parses the URL-encoded query string and
// fills in the `QueryParams` field.  Any existing values will
// be lost when this method is called.
func (r *Request) ParseQueryParameters() bool {
	r.QueryParams = nil
	indices := make(map[string]int)
	query := r.URL.RawQuery
	ok := true

	for i := 0; query != ""; {
		key := query
		if idx := strings.IndexAny(key, "&;"); idx >= 0 {
			key, query = key[:idx], key[idx+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if idx := strings.Index(key, "="); idx >= 0 {
			key, value = key[:idx], key[idx+1:]
		}

		key1, err := url.QueryUnescape(key)
		if err == nil {
			key = key1
		}

		index, found := indices[key]
		if !found {
			indices[key] = i
			index = i
			r.QueryParams = append(r.QueryParams, &QueryParameter{
				Name:    key,
				Ordinal: i,
			})
		}
		parameter := r.QueryParams[index]

		if err != nil {
			parameter.Err = InvalidParameterNameEscape.Append(err.Error())
			ok = false
		}

		value1, err := url.QueryUnescape(value)
		if err == nil {
			value = value1
		} else if parameter.Err == nil {
			parameter.Err = InvalidParameterValueEscape.Append(err.Error())
			ok = false
		}

		parameter.Values = append(parameter.Values, value)
		i++
	}

	return ok
}

func (r *Request) QueryParamExists(name string) bool {
	for _, p := range r.QueryParams {
		if p.Name == name {
			return true
		}
	}

	return false
}

func (r *Request) PathParamExists(name string) bool {
	for _, p := range r.PathParams {
		if p.Name == name {
			return true
		}
	}

	return false
}

type byName []*QueryParameter

func (p byName) Len() int           { return len(p) }
func (p byName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p byName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// ValidateQueryParameters validates the values in `QueryParams`
// with the provided fields.
// Validation errors are assigned to the problematic parameter
// and placeholder parameter instances are created for missing
// required or unknown parameters.
// If a validator panics an error will be returned in `err`.
// The `malformed` and `unknown` return values indicate if any
// paramters fail validation or do not match a field,
// respectively.
func (r *Request) ValidateQueryParameters(fields []Field) (malformed bool, unknown bool, err merry.Error) {
	params := make([]*QueryParameter, len(r.QueryParams))
	copy(params, r.QueryParams)
	sort.Sort(byName(params))

	for _, field := range fields {
		var found bool
		for j, param := range params {
			if param.Err != nil {
				malformed = true
			}
			if field.Match(param.Name) {
				found = true
				params = append(params[:j], params[j+1:]...)
				if param.Err != nil {
					break
				}

				if vErr := field.Validate(param.Values, &err); vErr != nil {
					param.Err = MalformedQueryParamter.Append(vErr.Error())
					malformed = true
				}
				if err != nil {
					return
				}
				break
			}
		}

		if !found {
			if field.Default != "" {
				r.QueryParams = append(r.QueryParams, &QueryParameter{
					Name:    field.Name,
					Ordinal: -1,
					Values:  []string{field.Default},
				})
			} else if field.Required {
				r.QueryParams = append(r.QueryParams, &QueryParameter{
					Name:    field.Name,
					Ordinal: -1,
					Err:     MissingQueryParamter,
				})
				malformed = true
			}
		}
	}

	for _, param := range params {
		param.Err = UnknownQueryParamter
		unknown = true
	}

	return
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
