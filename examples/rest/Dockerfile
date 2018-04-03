FROM alpine:3.7

RUN apk --no-cache --update add go musl-dev

COPY ./ /root/go/src/github.com/percolate/shisa

WORKDIR /root/go/src/github.com/percolate/shisa/examples/rest

RUN go build -ldflags '-linkmode external -extldflags "-static" -s' -o /root/go/bin/rest github.com/percolate/shisa/examples/rest

FROM alpine:3.7

COPY --from=0 /root/go/bin/rest rest

EXPOSE 9001

CMD ["./rest", "-addr=127.0.0.1:9001"]
