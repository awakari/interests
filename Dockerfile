FROM golang:1.23.2-alpine3.20 AS builder
WORKDIR /go/src/interests
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM scratch
COPY --from=builder /go/src/interests/interests /bin/interests
ENTRYPOINT ["/bin/interests"]
