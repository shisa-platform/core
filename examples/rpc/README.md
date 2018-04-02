# RPC Web Service

This is a sample of a "modern" RPC web service that implements the
Hello service.  The IdP service is required to be running and
accessible.

## Running

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

