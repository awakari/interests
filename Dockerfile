FROM golang:1.24.1-alpine3.21 AS builder
WORKDIR /go/src/interests
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM scratch
COPY --from=builder /go/src/interests/interests /bin/interests
ENTRYPOINT ["/bin/interests"]
