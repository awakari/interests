.PHONY: test clean
default: build

BINARY_FILE_NAME=synapse
COVERAGE_FILE_NAME=cover.out

vet:
	go vet

test: vet
	go test -race -cover -coverprofile=${COVERAGE_FILE_NAME} ./...
	cat ${COVERAGE_FILE_NAME} | grep -v mock > ${COVERAGE_FILE_NAME}.tmp
	mv -f ${COVERAGE_FILE_NAME}.tmp ${COVERAGE_FILE_NAME}
	go tool cover -func=${COVERAGE_FILE_NAME} | grep -Po '^total\:\h+\(statements\)\h+\K[\d]+' > cover.tmp
	./cover.sh

build:
	CGO_ENABLED=0 GOOS=linux GOARCH= GOARM= go build -o ${BINARY_FILE_NAME} cmd/main.go
	chmod ugo+x ${BINARY_FILE_NAME}

docker: build
	docker build -t cloud-message-bus/synapse .

rundb:
	docker run \
		-d \
		--name synapse-db \
		--network synapse-net \
		-p 27017:27017 \
		-v data:/data/db \
		mongo:latest \
		--nojournal

run: # TODO add cmd line args
	docker run \
		-d \
		--name synapse \
		--network synapse-net \
		-p 8080:8080 \
		--expose 8081 \
		cloud-message-bus/synapse

clean:
	go clean
	rm -f ${BINARY_FILE_NAME} ${COVERAGE_FILE_NAME}
