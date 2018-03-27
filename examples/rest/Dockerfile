FROM shisa-examples/base:demo

WORKDIR /root/go/src/github.com/percolate/shisa/examples/rest

RUN go build -o /root/go/bin/rest github.com/percolate/shisa/examples/rest

FROM alpine:3.7

COPY --from=0 /root/go/bin/rest rest

EXPOSE 9501

CMD ["./rest", "-addr=127.0.0.1:9501"]
