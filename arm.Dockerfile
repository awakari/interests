FROM golang:1.24.5-alpine3.22 AS builder
WORKDIR /go/src/interests
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build-arm64

FROM --platform=linux/arm64 scratch
COPY --from=builder /go/src/interests/interests /bin/interests
ENTRYPOINT ["/bin/interests"]
