GOPATH = $(shell go env GOPATH)

SHELL := /bin/bash -o pipefail

.PHONY: all
all: clean format lint compile

.PHONY: clean
clean:
	rm -rf target

target:
	mkdir target

.PHONY: format
format:
	goimports -w -local github.com/tomcz/example-grpc .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: tidy
tidy:
	go mod tidy -compat=1.21

.local/bin/protoc:
	mkdir .local
	curl -L -o .local/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip
	unzip .local/protoc.zip -d .local
	chmod +x .local/bin/protoc
	rm .local/protoc.zip

.local/googleapis:
	git clone --depth=1 https://github.com/googleapis/googleapis.git .local/googleapis

.PHONY: genproto
genproto: .local/bin/protoc .local/googleapis
	go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc
	.local/bin/protoc -I . \
		--go_out . --go_opt paths=source_relative \
		--go-grpc_out . --go-grpc_opt paths=source_relative \
		--grpc-gateway_out . --grpc-gateway_opt paths=source_relative \
		 -I.local/googleapis \
		api/service.proto

.PHONY: compile
compile: compile-server compile-client compile-certs

.PHONY: compile-server
compile-server: target
	go build -o target/example-server ./cmd/example-server/...

.PHONY: compile-client
compile-client: target
	go build -o target/example-client ./cmd/example-client/...

.PHONY: compile-certs
compile-certs: target
	go build -o target/example-certs ./cmd/example-certs/...
	./target/example-certs

# ========================================================================================
# Server variants
# ========================================================================================

.PHONY: run-server
run-server: compile-server
	./target/example-server -tokens "alice:wibble"

.PHONY: run-server-mtls
run-server-mtls: compile-server
	./target/example-server -tokens "alice:wibble" -domains "alice.example.com"

.PHONY: run-all-tests
run-all-tests: run-client-tests run-curl-tests run-grpcurl-tests

# ========================================================================================
# Custom gRPC client
# ========================================================================================

.PHONY: run-client
run-client: compile-client
	./target/example-client -token wibble -msg "G'day"

.PHONY: run-client-alice
run-client-alice: compile-client
	./target/example-client -alice -msg "Tea?"

.PHONY: run-client-bob
run-client-bob: compile-client
	-./target/example-client -bob -msg "Coffee?"

.PHONY: run-client-tests
run-client-tests: run-client run-client-alice run-client-bob

# ========================================================================================
# Plain HTTP client: curl
# Download jq from: https://stedolan.github.io/jq/download/
# ========================================================================================

.PHONY: run-curl
run-curl:
	curl --silent --show-error --fail \
		--cacert target/ca.crt \
		-H 'Content-Type: application/json' \
		-H 'Authorization: Bearer wibble' \
		-d '{"message": "hello"}' \
		https://localhost:8443/v1/example/echo | jq '.'

.PHONY: run-curl-alice
run-curl-alice:
	curl --silent --show-error --fail \
		--cacert target/ca.crt \
		--cert target/alice.crt \
		--key target/alice.key \
		-H 'Content-Type: application/json' \
		-d '{"message": "Wine?"}' \
		https://localhost:8443/v1/example/echo | jq '.'

.PHONY: run-curl-bob
run-curl-bob:
	-curl --silent --show-error --fail \
		--cacert target/ca.crt \
		--cert target/bob.crt \
 		--key target/bob.key \
		-H 'Content-Type: application/json' \
		-d '{"message": "Whiskey?"}' \
		https://localhost:8443/v1/example/echo

.PHONY: run-curl-tests
run-curl-tests: run-curl run-curl-alice run-curl-bob

# ========================================================================================
# Third-party gRPC client: grpcurl
# Download from: https://github.com/fullstorydev/grpcurl
# ========================================================================================

.PHONY: run-grpcurl
run-grpcurl:
	grpcurl -servername server.example.com \
		-cacert target/ca.crt \
		-H 'authorization: bearer wibble' \
		-d '{"message":"hola"}' \
		localhost:8000 example.service.Example/Echo

.PHONY: run-grpcurl-alice
run-grpcurl-alice:
	grpcurl -servername server.example.com \
		-cacert target/ca.crt \
		-cert target/alice.crt \
		-key target/alice.key \
		-d '{"message":"Gin?"}' \
		localhost:8000 example.service.Example/Echo

.PHONY: run-grpcurl-bob
run-grpcurl-bob:
	-grpcurl -servername server.example.com \
		-cacert target/ca.crt \
		-cert target/bob.crt \
		-key target/bob.key \
		-d '{"message":"Vodka?"}' \
		localhost:8000 example.service.Example/Echo

.PHONY: run-grpcurl-tests
run-grpcurl-tests: run-grpcurl run-grpcurl-alice run-grpcurl-bob
