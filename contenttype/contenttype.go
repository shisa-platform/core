package contenttype

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ApplicationMediaType    = "application"
	OctetStreamMediaSubtype = "octet-stream"
	TextMediaType           = "text"
	PlainMediaSubtype       = "plain"
	JsonMediaSubtype        = "json"
	XmlMediaSubtype         = "xml"
	DefaultMediaType        = ApplicationMediaType
	DefaultMediaSubtype     = JsonMediaSubtype
)

var DefaultContentType = &ContentType{
	MediaType:    DefaultMediaType,
	MediaSubtype: DefaultMediaSubtype,
}

var TextPlainContentType = &ContentType{
	MediaType:    TextMediaType,
	MediaSubtype: PlainMediaSubtype,
}

var ApplicationJsonContentType = &ContentType{
	MediaType:    ApplicationMediaType,
	MediaSubtype: JsonMediaSubtype,
}

var ApplicationXmlContentType = &ContentType{
	MediaType:    ApplicationMediaType,
	MediaSubtype: XmlMediaSubtype,
}

type ContentType struct {
	MediaType    string
	MediaSubtype string
}

func New(mediaType, mediaSubtype string) *ContentType {
	return &ContentType{
		MediaType:    mediaType,
		MediaSubtype: mediaSubtype,
	}
}

func NewFromString(ct string) (*ContentType, error) {
	split := strings.Split(ct, "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("Malformed content type: %s", ct)
	}
	return New(split[0], split[1]), nil
}

func (ct *ContentType) String() string {
	return fmt.Sprintf("%s/%s", ct.MediaType, ct.MediaSubtype)
}

func (ct *ContentType) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(ct.String())), nil
}

type ContentTypeMap map[ContentType]interface{}

func (c ContentTypeMap) Get(key ContentType) (interface{}, bool) {
	if value, ok := c[key]; ok {
		return value, ok
	}

	wildcard := ContentType{MediaType: key.MediaType, MediaSubtype: "*"}
	if value, ok := c[wildcard]; ok {
		return value, ok
	}
	return nil, false
}
