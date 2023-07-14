FROM golang:1.20.6-alpine3.18 AS builder
WORKDIR /go/src/subscriptions
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM scratch
COPY --from=builder /go/src/subscriptions/subscriptions /bin/subscriptions
ENTRYPOINT ["/bin/subscriptions"]
