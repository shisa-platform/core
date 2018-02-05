# Example Identity Provider Service

This is a sample of an Identity Provider (IdP) service to offload
Authentication (AuthN) and Authorization (AuthZ) decisions to a
separate system.

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
    string Value = 2;
}

message User {
    string Ident = 1;
    string Name = 2;
    string Pass = 3;
}

service Idp {
    rpc AuthenticateToken(Message) returns (string);
    rpc FindUser(Message) returns (User);
    rpc Healthcheck(string) returns (bool);
}
```

