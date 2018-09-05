package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
)

var (
	expectedRoute  = "/foo"
	expectedPolicy = Policy{AllowTrailingSlashRedirects: true}
	emptyPolicy    = Policy{}
)

func testHandler(context.Context, *httpx.Request) httpx.Response {
	return nil
}

func TestGetEndpoint(t *testing.T) {
	cut := GetEndpoint(expectedRoute, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Get)
	assert.Len(t, cut.Get.Handlers, 1)
	assert.Equal(t, emptyPolicy, cut.Get.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Put)
	assert.Nil(t, cut.Post)
	assert.Nil(t, cut.Patch)
	assert.Nil(t, cut.Delete)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestGetEndpointWithPolicy(t *testing.T) {
	cut := GetEndpointWithPolicy(expectedRoute, expectedPolicy, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Get)
	assert.Len(t, cut.Get.Handlers, 1)
	assert.Equal(t, expectedPolicy, cut.Get.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Put)
	assert.Nil(t, cut.Post)
	assert.Nil(t, cut.Patch)
	assert.Nil(t, cut.Delete)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestPutEndpoint(t *testing.T) {
	cut := PutEndpoint(expectedRoute, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Put)
	assert.Len(t, cut.Put.Handlers, 1)
	assert.Equal(t, emptyPolicy, cut.Put.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Get)
	assert.Nil(t, cut.Post)
	assert.Nil(t, cut.Patch)
	assert.Nil(t, cut.Delete)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestPutEndpointWithPolicy(t *testing.T) {
	cut := PutEndpointWithPolicy(expectedRoute, expectedPolicy, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Put)
	assert.Len(t, cut.Put.Handlers, 1)
	assert.Equal(t, expectedPolicy, cut.Put.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Get)
	assert.Nil(t, cut.Post)
	assert.Nil(t, cut.Patch)
	assert.Nil(t, cut.Delete)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestPostEndpoint(t *testing.T) {
	cut := PostEndpoint(expectedRoute, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Post)
	assert.Len(t, cut.Post.Handlers, 1)
	assert.Equal(t, emptyPolicy, cut.Post.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Get)
	assert.Nil(t, cut.Put)
	assert.Nil(t, cut.Patch)
	assert.Nil(t, cut.Delete)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestPostEndpointWithPolicy(t *testing.T) {
	cut := PostEndpointWithPolicy(expectedRoute, expectedPolicy, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Post)
	assert.Len(t, cut.Post.Handlers, 1)
	assert.Equal(t, expectedPolicy, cut.Post.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Get)
	assert.Nil(t, cut.Put)
	assert.Nil(t, cut.Patch)
	assert.Nil(t, cut.Delete)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestPatchEndpoint(t *testing.T) {
	cut := PatchEndpoint(expectedRoute, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Patch)
	assert.Len(t, cut.Patch.Handlers, 1)
	assert.Equal(t, emptyPolicy, cut.Patch.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Get)
	assert.Nil(t, cut.Put)
	assert.Nil(t, cut.Post)
	assert.Nil(t, cut.Delete)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestPatchEndpointWithPolicy(t *testing.T) {
	cut := PatchEndpointWithPolicy(expectedRoute, expectedPolicy, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Patch)
	assert.Len(t, cut.Patch.Handlers, 1)
	assert.Equal(t, expectedPolicy, cut.Patch.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Get)
	assert.Nil(t, cut.Put)
	assert.Nil(t, cut.Post)
	assert.Nil(t, cut.Delete)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestDeleteEndpoint(t *testing.T) {
	cut := DeleteEndpoint(expectedRoute, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Delete)
	assert.Len(t, cut.Delete.Handlers, 1)
	assert.Equal(t, emptyPolicy, cut.Delete.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Get)
	assert.Nil(t, cut.Put)
	assert.Nil(t, cut.Post)
	assert.Nil(t, cut.Patch)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestDeleteEndpointWithPolicy(t *testing.T) {
	cut := DeleteEndpointWithPolicy(expectedRoute, expectedPolicy, testHandler)
	assert.Equal(t, expectedRoute, cut.Route)
	assert.NotNil(t, cut.Delete)
	assert.Len(t, cut.Delete.Handlers, 1)
	assert.Equal(t, expectedPolicy, cut.Delete.Policy)

	assert.Nil(t, cut.Head)
	assert.Nil(t, cut.Get)
	assert.Nil(t, cut.Put)
	assert.Nil(t, cut.Post)
	assert.Nil(t, cut.Patch)
	assert.Nil(t, cut.Connect)
	assert.Nil(t, cut.Options)
	assert.Nil(t, cut.Trace)
}

func TestEndpointExpvarStringExerciseComma(t *testing.T) {
	cut := GetEndpointWithPolicy(expectedRoute, expectedPolicy, testHandler)
	cut.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "thing", Required: true},
	}
	cut.Put = &Pipeline{
		Handlers: []httpx.Handler{testHandler},
	}

	val := cut.String()

	assert.NotEmpty(t, val)
	expectedJSON := `{
  "GET": {
    "Policy": {
      "AllowTrailingSlashRedirects": true
    },
    "Handlers": 1,
    "QuerySchemas": [
      {"Name": "thing", "Regex": null, "Required": true}
    ]
  },
  "PUT": {
    "Policy": {},
    "Handlers": 1
  }
}`
	assert.JSONEq(t, expectedJSON, val)
}

func TestEndpointExpvarString(t *testing.T) {
	cut := GetEndpointWithPolicy(expectedRoute, expectedPolicy, testHandler)
	cut.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "thing", Required: true},
	}
	cut.Head = &Pipeline{
		Handlers: []httpx.Handler{testHandler},
	}
	cut.Put = &Pipeline{
		Handlers: []httpx.Handler{testHandler},
	}
	cut.Post = &Pipeline{
		Handlers: []httpx.Handler{testHandler},
	}
	cut.Patch = &Pipeline{
		Handlers: []httpx.Handler{testHandler},
	}
	cut.Delete = &Pipeline{
		Handlers: []httpx.Handler{testHandler},
	}
	cut.Connect = &Pipeline{
		Handlers: []httpx.Handler{testHandler},
	}
	cut.Options = &Pipeline{
		Handlers: []httpx.Handler{testHandler},
	}
	cut.Trace = &Pipeline{
		Handlers: []httpx.Handler{testHandler},
	}

	val := cut.String()

	assert.NotEmpty(t, val)
	expectedJSON := `{
  "HEAD": {
    "Policy": {},
    "Handlers": 1
  },
  "GET": {
    "Policy": {
      "AllowTrailingSlashRedirects": true
    },
    "Handlers": 1,
    "QuerySchemas": [
      {"Name": "thing", "Regex": null, "Required": true}
    ]
  },
  "PUT": {
    "Policy": {},
    "Handlers": 1
  },
  "POST": {
    "Policy": {},
    "Handlers": 1
  },
  "PATCH": {
    "Policy": {},
    "Handlers": 1
  },
  "DELETE": {
    "Policy": {},
    "Handlers": 1
  },
  "CONNECT": {
    "Policy": {},
    "Handlers": 1
  },
  "OPTIONS": {
    "Policy": {},
    "Handlers": 1
  },
  "TRACE": {
    "Policy": {},
    "Handlers": 1
  }
}`
	assert.JSONEq(t, expectedJSON, val)
}
