FROM shisa-examples/base:demo

WORKDIR /root/go/src/github.com/percolate/shisa/examples/rpc

RUN go build -o /root/go/bin/rpc github.com/percolate/shisa/examples/rpc

FROM alpine:3.7

COPY --from=0 /root/go/bin/rpc rpc

EXPOSE 9401

CMD ["./rpc", "-addr=127.0.0.1:9401"]
