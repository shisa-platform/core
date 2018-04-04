# Examples

The packages in this directory are example illustrating how to use
Shisa.

## Architecture

![design](examples.png)

## Services

- [`gw`](gw) - A Shisa implemenation talking to the `rest` and `rpc` services.
- [`rest`](rest) - A "legacy" RESTful web service implementing the "Goodbye" service
- [`rpc`](rpc) - A "modern" RPC service implementing the "Hello" service
- [`idp`](idp) - An Identity Provider service used by the other services.

## Building

You can build the services as docker images by running the
following from the project root (requires docker + docker-compose):

    make docker

## Running

Run the following command at the project root:

    docker-compose -f examples/docker-compose.yml up

There are now an instance of the `gw` service bound to the host port
`9001`.  Consul is bound to host port `8500` and its UI is available
via web browser at `127.0.0.1:8500/ui/`.

### API Gateway

The `healthcheck` and `debug` endpoints require authentication by the
user `Admin` with the password: `password`.  The `api/greeting` and 
`api/farewell` endpoints can be accessed by the `Admin` user or  the
user `Boss` with password `password`.  Refer to the [RAML specification](examples/gw/README.md#endpoints)
for details sabout the `api/greeting` and `api/farewell` endpoints.

- Healthcheck - <http://localhost:9003/healthcheck>
- Debug Vars - <http://localhost:9002/debug/vars>
- Greeting Endpoint - <http://localhost:9001/api/greeting>
- Farewell Endpoint - <http://localhost:9001/api/farewell>
