FROM shisa/examples/base

WORKDIR /root/go/src/github.com/percolate/shisa/examples/idp

RUN go build -ldflags '-linkmode external -extldflags "-static" -s' -o /root/go/bin/idp github.com/percolate/shisa/examples/idp

FROM alpine:3.7

COPY --from=0 /root/go/bin/idp idp

EXPOSE 9601

CMD ["./idp", "-addr=127.0.0.1:9601"]
