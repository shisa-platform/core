package service

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// Pipeline is a chain of handlers to be invoked in order on a
// request.  The first non-nil response will be returned to the
// user agent.  If no response is produced an Internal Service
// Error handler will be invoked.
type Pipeline struct {
	Policy      Policy    // customizes automated behavior
	Handlers    []Handler // the pipline steps, minimum one
	QueryFields []Field   // optional query parameter validation
}

func (p Pipeline) jsonify(buf *bytes.Buffer) {
	enc := json.NewEncoder(buf)
	buf.WriteString("{\"Policy\":")
	enc.Encode(p.Policy)
	buf.WriteString(",\"Handlers\":")
	buf.WriteString(strconv.Itoa(len(p.Handlers)))
	if len(p.QueryFields) != 0 {
		buf.WriteString(",\"QueryFields\":")
		enc.Encode(p.QueryFields)
	}
	buf.WriteByte('}')
}
