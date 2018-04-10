# Examples

The packages in this directory illustrate how to use Shisa.

## Architecture

![design](examples.png)

## Services

- [`gw`](gw) - A Shisa implementation talking to the `rest` and `rpc` services.
- [`rpc`](rpc) - A "modern" RPC service implementing the "Hello" service
- [`rest`](rest) - A "legacy" RESTful web service implementing the "Goodbye" service
- [`idp`](idp) - An Identity Provider service used by the other services.

### Service Discovery

Shisa has built-in support for service discovery and load balancing,
and the examples use provided [Consul](https://www.consul.io/) client.
All of the example services register themselves so that they can be
automatically discovered by their downstream dependencies.  Each
example service also registers a health check so Consul can track the
availability of all service instances.

There is only a single instance of Consul in the example cluster,
unlike a typical production deployment which would have a Consul agent
on each node and a cluster of consul servers elsewhere.  This requires
configuring instance ids in Consul which is usually not necessary in a
production environment.

The Consul web application is available at: <http://localhost:8500/ui/>

### Distributed Tracing

Shisa has built-in support for [OpenTracing](http://opentracing.io/)
and the examples use [Jaeger](https://www.jaegertracing.io/) to
capture spans emitted during each API request.  The examples configure
the Jaeger agent to capture all spans rather that a the typical
production setup of a sampling.

The Jaeger web application is available at: <http://localhost:16686/>

## Building

Run the following command at the project root:

    make docker

## Running

Run the following command at the project root:

    docker-compose -f examples/docker-compose.yml up

There are now an instance of the `gw` service bound to the host port
`9001`.

## API Gateway

The `healthcheck` and `debug` endpoints require authentication by the
user `Admin` with the password: `password`.  The `api/greeting` and
`api/farewell` endpoints can be accessed by the `Admin` user or  the
user `Boss` with password `password`.  Refer to the [RAML specification](examples/gw/README.md#endpoints)
for details about the `api/greeting` and `api/farewell` endpoints.

- Greeting Endpoint - <http://localhost:9001/api/greeting>
- Farewell Endpoint - <http://localhost:9001/api/farewell>
- Debug Vars - <http://localhost:9002/debug/vars>
- Health Check - <http://localhost:9003/healthcheck>
