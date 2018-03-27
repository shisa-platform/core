FROM shisa-examples/base:demo

WORKDIR /root/go/src/github.com/percolate/shisa/examples/gw

RUN go build -o /root/go/bin/gw github.com/percolate/shisa/examples/gw

FROM alpine:3.7

COPY --from=0 /root/go/bin/gw gw

EXPOSE 9001

CMD ["./gw", "-addr=127.0.0.1:9001", "-debugaddr=127.0.0.1:9002", "-healthcheckaddr=127.0.0.1:9003"]
