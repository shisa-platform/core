package sd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"testing"

	"github.com/ansel1/merry"
	consul "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

const (
	testServiceName = "service"
	testAddr        = "127.0.0.1:9000"
)

func TestNewConsul(t *testing.T) {
	ac := &consul.Client{}

	c := NewConsul(ac)

	assert.NotNil(t, c)
	assert.NotNil(t, c.agent)
	assert.NotNil(t, c.health)
}

func TestConsulSDRegisterBadAddr(t *testing.T) {
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	e := c.Register(testServiceName, "bad addr")

	assert.Error(t, e)
	freg.AssertServiceRegisterNotCalled(t)
}

func TestConsulSDRegisterBadPort(t *testing.T) {
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	e := c.Register(testServiceName, "127.0.0.1:badport")

	assert.Error(t, e)
	freg.AssertServiceRegisterNotCalled(t)
}

func TestConsulSDRegisterServiceRegisterError(t *testing.T) {
	terr := merry.New("service registry error")
	freg := &FakeconsulRegistry{
		ServiceRegisterHook: func(*consul.AgentServiceRegistration) error {
			return terr
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	e := c.Register(testServiceName, testAddr)

	assert.Error(t, e)
	assert.Equal(t, terr, e)
	freg.AssertServiceRegisterCalledOnce(t)
}

func TestConsulSDRegisterSuccess(t *testing.T) {
	freg := &FakeconsulRegistry{
		ServiceRegisterHook: func(*consul.AgentServiceRegistration) error {
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	e := c.Register(testServiceName, testAddr)

	assert.Nil(t, e)
	freg.AssertServiceRegisterCalledOnce(t)
}

func TestConsulSDDeregisterError(t *testing.T) {
	terr := merry.New("service deregister error")
	freg := &FakeconsulRegistry{
		ServiceDeregisterHook: func(string) error {
			return terr
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	e := c.Deregister(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, terr, e)
	freg.AssertServiceDeregisterCalledOnceWith(t, testServiceName)
}

func TestConsulSDDeregisterSuccess(t *testing.T) {
	freg := &FakeconsulRegistry{
		ServiceDeregisterHook: func(string) error {
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	e := c.Deregister(testServiceName)

	assert.Nil(t, e)
	freg.AssertServiceDeregisterCalledOnceWith(t, testServiceName)
}

func TestConsulSDResolveError(t *testing.T) {
	terr := merry.New("service resolve error")
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{
		ServiceHook: func(string, string, bool, *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
			return []*consul.ServiceEntry{}, nil, terr
		},
	}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	n, e := c.Resolve(testServiceName)

	assert.Empty(t, n)
	assert.Error(t, e)
	fres.AssertServiceCalledOnceWith(t, testServiceName, "", true, nil)
}

func TestConsulSDResolveNodeAddress(t *testing.T) {
	host, sport, e := net.SplitHostPort(testAddr)
	if e != nil {
		t.Error(e)
	}
	port, e := strconv.Atoi(sport)
	if e != nil {
		t.Error(e)
	}

	n := &consul.Node{Address: host}
	s := &consul.AgentService{Port: port}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{
		ServiceHook: func(string, string, bool, *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
			return []*consul.ServiceEntry{{Node: n, Service: s}}, nil, nil
		},
	}
	exp := []string{testAddr}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	res, e := c.Resolve(testServiceName)

	assert.Equal(t, exp, res)
	assert.Nil(t, e)
	fres.AssertServiceCalledOnceWith(t, testServiceName, "", true, nil)
}

func TestConsulSDResolveServiceAddress(t *testing.T) {
	host, sport, e := net.SplitHostPort(testAddr)
	if e != nil {
		t.Error(e)
	}
	port, e := strconv.Atoi(sport)
	if e != nil {
		t.Error(e)
	}

	n := &consul.Node{}
	s := &consul.AgentService{Address: host, Port: port}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{
		ServiceHook: func(string, string, bool, *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
			return []*consul.ServiceEntry{{Node: n, Service: s}}, nil, nil
		},
	}
	exp := []string{testAddr}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	res, e := c.Resolve(testServiceName)

	assert.Equal(t, exp, res)
	assert.Nil(t, e)
	fres.AssertServiceCalledOnceWith(t, testServiceName, "", true, nil)
}

func TestConsulSDAddCheckGRPC(t *testing.T) {
	surl := "grpc://127.0.0.1:9000?interval=5s"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{
		CheckRegisterHook: func(*consul.AgentCheckRegistration) error {
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Nil(t, err)
	freg.AssertCheckRegisterCalledOnce(t)
}

func TestConsulSDAddCheckGRPCMissingHost(t *testing.T) {
	surl := "grpc://" // missing `host`
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Error(t, err)
	freg.AssertCheckDeregisterNotCalled(t)
}

func TestConsulSDAddCheckGRPCMissingInterval(t *testing.T) {
	surl := "grpc://localhost:9000" // missing `interval`
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Error(t, err)
	freg.AssertCheckDeregisterNotCalled(t)
}

func TestConsulSDAddCheckDockerMissingKey(t *testing.T) {
	surl := "docker://" // missing keys
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Error(t, err)
	freg.AssertCheckDeregisterNotCalled(t)
}

func TestConsulSDAddCheckDocker(t *testing.T) {
	surl := "docker://127.0.0.1:9000?dockercontainerid=d858g8ergj&args=\"run command\"&interval=5s"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{
		CheckRegisterHook: func(*consul.AgentCheckRegistration) error {
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Nil(t, err)
	freg.AssertCheckRegisterCalledOnce(t)
}

func TestConsulSDAddCheckScriptMissingKey(t *testing.T) {
	surl := "script://" // missing keys
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Error(t, err)
	freg.AssertCheckDeregisterNotCalled(t)
}

func TestConsulSDAddCheckScript(t *testing.T) {
	surl := "script://127.0.0.1:9000?args=\"run command\"&interval=5s"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{
		CheckRegisterHook: func(*consul.AgentCheckRegistration) error {
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Nil(t, err)
	freg.AssertCheckRegisterCalledOnce(t)
}

func TestConsulSDAddCheckTTL(t *testing.T) {
	surl := "ttl://127.0.0.1:9000?ttl=5s"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{
		CheckRegisterHook: func(c *consul.AgentCheckRegistration) error {
			assert.NotNil(t, c.TTL)
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Nil(t, err)
	freg.AssertCheckRegisterCalledOnce(t)
}

func TestConsulSDAddCheckTCP(t *testing.T) {
	surl := "tcp://127.0.0.1:9000?interval=5s"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{
		CheckRegisterHook: func(*consul.AgentCheckRegistration) error {
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Nil(t, err)
	freg.AssertCheckRegisterCalledOnce(t)
}

func TestConsulSDAddCheckTCPMissingHost(t *testing.T) {
	surl := "tcp://" // missing host
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Error(t, err)
	freg.AssertCheckDeregisterNotCalled(t)
}

func TestConsulSDAddCheckTCPMissingKey(t *testing.T) {
	surl := "tcp://127.0.0.1:9000" // missing interval
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Error(t, err)
	freg.AssertCheckDeregisterNotCalled(t)
}

func TestConsulSDAddCheckHTTP(t *testing.T) {
	surl := "http://127.0.0.1:9000?interval=5s"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{
		CheckRegisterHook: func(*consul.AgentCheckRegistration) error {
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Nil(t, err)
	freg.AssertCheckRegisterCalledOnce(t)
}

func TestConsulSDAddCheckHTTPS(t *testing.T) {
	surl := "https://127.0.0.1:9000?interval=5s"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{
		CheckRegisterHook: func(*consul.AgentCheckRegistration) error {
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Nil(t, err)
	freg.AssertCheckRegisterCalledOnce(t)
}

func TestConsulSDAddCheckHTTPSMissingKey(t *testing.T) {
	surl := "https://127.0.0.1:9000"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Error(t, err)
	freg.AssertCheckDeregisterNotCalled(t)
}

func TestConsulSDAddCheckHTTPSMissingURL(t *testing.T) {
	surl := "https://"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Error(t, err)
	freg.AssertCheckDeregisterNotCalled(t)
}

func TestConsulSDAddCheckCheckRegisterError(t *testing.T) {
	terr := merry.New("error in consul registration")
	surl := "https://127.0.0.1:9000?interval=5s"
	u, e := url.Parse(surl)
	if e != nil {
		t.Error(e)
	}
	freg := &FakeconsulRegistry{
		CheckRegisterHook: func(*consul.AgentCheckRegistration) error {
			return terr
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	err := c.AddCheck(testServiceName, u)

	assert.Error(t, err)
	assert.Equal(t, terr, err)
	freg.AssertCheckRegisterCalledOnce(t)

}

func TestConsulSDRemoveChecksError(t *testing.T) {
	hcname := fmt.Sprintf("%s-healthcheck", testServiceName)
	terr := merry.New("service check deregister error")
	freg := &FakeconsulRegistry{
		CheckDeregisterHook: func(string) error {
			return terr
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	e := c.RemoveChecks(testServiceName)

	assert.Error(t, e)
	assert.Equal(t, terr, e)
	freg.AssertCheckDeregisterCalledOnceWith(t, hcname)
}

func TestConsulSDRemoveChecksSuccess(t *testing.T) {
	hcname := fmt.Sprintf("%s-healthcheck", testServiceName)
	freg := &FakeconsulRegistry{
		CheckDeregisterHook: func(string) error {
			return nil
		},
	}
	fres := &FakeconsulResolver{}
	c := consulSD{
		agent:  freg,
		health: fres,
	}

	e := c.RemoveChecks(testServiceName)

	assert.Nil(t, e)
	freg.AssertCheckDeregisterCalledOnceWith(t, hcname)
}

func TestPopQueryString(t *testing.T) {
	k := "k"
	sv := "v"
	v := make(url.Values, 1)
	v[k] = []string{sv}

	r := popQueryString(v, k)

	assert.Equal(t, sv, r)
}

func TestPopQueryStringNotFound(t *testing.T) {
	k := "k"
	v := make(url.Values, 1)

	r := popQueryString(v, k)

	assert.Equal(t, "", r)
}

func TestPopQueryBool(t *testing.T) {
	k := "k"
	b := true
	bv := []string{fmt.Sprint(b)}
	v := make(url.Values, 1)
	v[k] = bv

	r := popQueryBool(v, k)

	assert.Equal(t, b, r)
}

func TestPopQueryBoolNotFound(t *testing.T) {
	k := "k"
	v := make(url.Values, 1)

	r := popQueryBool(v, k)

	assert.Equal(t, false, r)
}

func TestPopQuerySlice(t *testing.T) {
	k := "k"
	sv := []string{"v1", "v2"}
	v := make(url.Values, 1)
	v[k] = sv

	r := popQuerySlice(v, k)

	assert.Equal(t, sv, r)
}

func TestPopQuerySliceNotFound(t *testing.T) {
	k := "k"
	v := make(url.Values, 1)

	r := popQuerySlice(v, k)

	assert.Equal(t, []string(nil), r)
}

func TestValidateKeyExists(t *testing.T) {
	k := "k"
	sv := []string{"v1", "v2"}
	v := make(url.Values, 1)
	v[k] = sv

	r := validateKeysExist(v, k)

	assert.Nil(t, r)
}

func TestValidateKeyExistsNotFound(t *testing.T) {
	k := "k"
	v := make(url.Values, 1)

	r := validateKeysExist(v, k)

	assert.Error(t, r)
}
