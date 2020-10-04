.PHONY: all clean format lint genproto
.PHONY: compile compile-server compile-client

GOPATH = $(shell go env GOPATH)
PACKAGES = $(shell go list ./... | grep -v vendor)

# default make shell is /bin/sh which does not support pipefail
SHELL=/bin/bash -eo pipefail

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

test-clients: run-client-tests run-curl-tests run-grpcurl-tests

# ========================================================================================
# Server variants
# ========================================================================================

run-server: compile-server
	./target/example-server

run-server-mtls: compile-server
	./target/example-server -mtls

# ========================================================================================
# Custom gRPC client
# ========================================================================================

run-client: compile-client
	./target/example-client -token wibble -msg "G'day"

run-client-alice: compile-client
	./target/example-client -alice -msg "Tea?"

run-client-bob: compile-client
	./target/example-client -bob -msg "Coffee?"

run-client-tests: run-client run-client-alice run-client-bob

# ========================================================================================
# Plain HTTP client: curl
# ========================================================================================

run-curl:
	curl --silent --show-error --fail --insecure  \
		-H 'Content-Type: application/json' \
		-H 'Authorization: Bearer letmein' \
		-d '{"message": "hello"}' https://localhost:8080/v1/example/echo | jq '.'

run-curl-alice:
	curl --silent --show-error --fail --insecure \
		--key pki/alice.key --cert pki/alice.crt \
		-H 'Content-Type: application/json' \
		-d '{"message": "Wine?"}' https://localhost:8080/v1/example/echo | jq '.'

run-curl-bob:
	curl --silent --show-error --fail --insecure \
		--key pki/bob.key --cert pki/bob.crt \
		-H 'Content-Type: application/json' \
		-d '{"message": "Whiskey?"}' https://localhost:8080/v1/example/echo | jq '.'

run-curl-tests: run-curl run-curl-alice run-curl-bob

# ========================================================================================
# Third-party gRPC client: grpcurl
# ========================================================================================

run-grpcurl:
	grpcurl -cacert pki/ca.crt -servername server.example.com \
		-d '{"message":"hola"}' -H 'authorization: bearer letmein' \
		localhost:8000 example.service.Example/Echo

run-grpcurl-alice:
	grpcurl -cacert pki/ca.crt -servername server.example.com \
		-cert pki/alice.crt -key pki/alice.key \
		-d '{"message":"Gin?"}' \
		localhost:8000 example.service.Example/Echo

run-grpcurl-bob:
	grpcurl -cacert pki/ca.crt -servername server.example.com \
		-cert pki/bob.crt -key pki/bob.key \
		-d '{"message":"Vodka?"}' \
		localhost:8000 example.service.Example/Echo

run-grpcurl-tests: run-grpcurl run-grpcurl-alice run-grpcurl-bob
