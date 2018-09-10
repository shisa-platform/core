package service

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/shisa-platform/core/httpx"
)

// Pipeline is a chain of handlers to be invoked in order on a
// request.  The first non-nil response will be returned to the
// user agent.  If no response is produced an Internal Service
// Error handler will be invoked.
type Pipeline struct {
	Policy       Policy                  // customizes automated behavior
	Handlers     []httpx.Handler         // the pipline steps, minimum one
	QuerySchemas []httpx.ParameterSchema // optional query parameter validation
}

func (p Pipeline) jsonify(buf *bytes.Buffer) {
	enc := json.NewEncoder(buf)
	buf.WriteString("{\"Policy\":")
	enc.Encode(p.Policy)
	buf.WriteString(",\"Handlers\":")
	buf.WriteString(strconv.Itoa(len(p.Handlers)))
	if len(p.QuerySchemas) != 0 {
		buf.WriteString(",\"QuerySchemas\":")
		enc.Encode(p.QuerySchemas)
	}
	buf.WriteByte('}')
}
