FROM golang:1.20.0-alpine3.17 AS builder
WORKDIR /go/src/subscriptions
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM alpine:3.17.0
COPY --from=builder /go/src/subscriptions/subscriptions /bin/subscriptions
ENTRYPOINT ["/bin/subscriptions"]
