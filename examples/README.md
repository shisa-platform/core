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

There are now 3 `gw` service instances bound to host ports `8001`, `8002` and
`8003`. Consul is bound to port `8500` and its UI is available via web browser
at `127.0.0.1:8500/ui/`.

### API Gateway

The `healthcheck` and `debug` endpoints require authentication by the
"admin" user: `Admin:password`.  The `api/greeting` and `api/farewell`
endpoints can be accessed by the "admin" user or `Boss:password`.

- Healthcheck - <http://localhost:9003/healthcheck>
- Debug Vars - <http://localhost:9002/debug/vars>
- Greeting Endpoint - <http://localhost:9001/api/greeting>
- Farewell Endpoint - <http://localhost:9001/api/farewell>
