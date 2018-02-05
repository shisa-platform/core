# RPC Web Service

This is a sample of a "modern" RPC web service that implements the
Hello service.  The IdP service is required to be running and
accessible.

## Running

This services needs to know the address of the IdP service as an
environment variable.  The value should be in the form of: `host:port`
where `host` is optional.  If `host` is provided it must be a hostname
or (v4 or v6) IP address.

    export IDP_SERVICE_ADDR=":9601"

See the [parent README.MD](../README.md) for instructions to run
the example services together.

## Endpoints

This example implements a RPC API using the Go standard library `rpc`
package.  The specification below is expressed in terms of gRPC:

``` protobuf
package idp;

message Message {
    string RequestID = 1;
    string UserID = 2;
    string Language = 3;
    string Name = 4;
}

service Idp {
    rpc Greeting(Message) returns (string);
    rpc Healthcheck(string) returns (bool);
}
```

