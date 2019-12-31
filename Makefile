.PHONY: all clean format lint genproto
.PHONY: compile compile-server compile-client
.PHONY: run-server run-client run-curl run-grpcurl

GOPATH = $(shell go env GOPATH)
PACKAGES = $(shell go list ./... | grep -v vendor)
BEARER_TOKEN ?= wibble

all: clean format lint compile

clean:
	rm -rf target

target:
	mkdir target

format:
	@echo "Running goimports ..."
	@goimports -w -local github.com/tomcz/example-grpc $(shell find . -type f -name '*.go' | grep -v '/vendor/')

lint:
	@echo "Running govet ..."
	@go vet ${PACKAGES}
	@echo "Running golint ..."
	@golint -set_exit_status ${PACKAGES}

genproto:
	protoc -Iapi/ -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:api \
		api/service.proto
	protoc -Iapi/ -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:api \
		api/service.proto

compile: compile-server compile-client

compile-server: target
	go build -o target/example-server ./cmd/example-server/...

compile-client: target
	go build -o target/example-client ./cmd/example-client/...

run-server: compile-server
	./target/example-server

run-server-mw: compile-server
	./target/example-server -middleware

run-server-ref: compile-server
	./target/example-server -reflection

run-client: compile-client
	./target/example-client -token ${BEARER_TOKEN} -msg "G'day"

run-curl:
	curl -s -H 'Content-Type: application/json' -H 'Authorization: Bearer ${BEARER_TOKEN}' \
		-d '{"message": "hello"}' http://localhost:8080/v1/example/echo | jq '.'

run-grpcurl:
	grpcurl -plaintext -d '{"message":"howdy"}' -H 'authorization: bearer ${BEARER_TOKEN}' \
		-import-path ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-import-path ./api \
		-proto api/service.proto  \
		localhost:8000 example.service.Example/Echo

run-grpcurl-ref:
	grpcurl -plaintext -d '{"message":"hola"}' -H 'authorization: bearer ${BEARER_TOKEN}' \
		localhost:8000 example.service.Example/Echo
