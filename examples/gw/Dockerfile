FROM shisa/examples/base

WORKDIR /root/go/src/github.com/percolate/shisa/examples/gw

RUN go build -ldflags '-linkmode external -extldflags "-static" -s' -o /root/go/bin/gw github.com/percolate/shisa/examples/gw

FROM alpine:3.7

COPY --from=0 /root/go/bin/gw gw

EXPOSE 9001 9002 9003

CMD ["./gw", "-addr=127.0.0.1:9001", "-debugaddr=127.0.0.1:9002", "-healthcheckaddr=127.0.0.1:9003"]
