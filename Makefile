.PHONY: all clean format lint genproto
.PHONY: compile compile-server compile-client
.PHONY: run-server run-grpc-client run-http-client

GOPATH = $(shell go env GOPATH)
PACKAGES = $(shell go list ./... | grep -v vendor)

all: clean format lint compile

clean:
	rm -rf target

target:
	mkdir target

format:
	goimports -w -local github.com/tomcz/example-grpc .

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

run-grpc-client: compile-client
	./target/example-client

run-http-client:
	curl -s -H 'Content-Type: application/json' -d '{"message": "hello"}' http://localhost:8080/v1/example/echo | jq '.'
