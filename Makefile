GOPATH = $(shell go env GOPATH)

SHELL := /bin/bash -o pipefail

ifeq "$(shell uname -o)" "Darwin"
ifeq "$(shell uname -m)" "x86_64"
JQ_URL := https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-macos-amd64
GRPCURL_URL := https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_osx_x86_64.tar.gz
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-osx-x86_64.zip
LINT_URL := https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-darwin-amd64.tar.gz
else
JQ_URL := https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-macos-arm64
GRPCURL_URL := https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_osx_arm64.tar.gz
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-osx-universal_binary.zip
LINT_URL := https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-darwin-arm64.tar.gz
endif
else
JQ_URL := https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux-amd64
GRPCURL_URL := https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_linux_x86_64.tar.gz
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip
LINT_URL := https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-linux-amd64.tar.gz
endif

.PHONY: all
all: clean format lint compile

.PHONY: clean
clean:
	rm -rf target

target:
	mkdir target

.PHONY: format
format:
ifeq ($(shell which goimports),)
	go install golang.org/x/tools/cmd/goimports
endif
	goimports -w -local github.com/tomcz/example-grpc .

.local/bin/golangci-lint:
	mkdir -p .local/bin
	curl -s -L -o .local/golangci.tar.gz ${LINT_URL}
	tar -xzf .local/golangci.tar.gz --strip-components=1 -C .local/bin
	chmod +x .local/bin/golangci-lint
	rm .local/golangci.tar.gz

.PHONY: lint
lint: .local/bin/golangci-lint
	.local/bin/golangci-lint run

.PHONY: tidy
tidy:
	go mod tidy -compat=1.21

.local/bin/protoc:
	mkdir -p .local/bin
	curl -s -L -o .local/protoc.zip ${PROTOC_URL}
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
		 -I .local/googleapis \
		api/service.proto

.PHONY: compile
compile: target/example-server target/example-client target/example-certs

target/example-server: target
	go build -o target/example-server ./cmd/example-server/...

target/example-client: target
	go build -o target/example-client ./cmd/example-client/...

target/example-certs: target
	go build -o target/example-certs ./cmd/example-certs/...

# ========================================================================================
# Server and tests
# ========================================================================================

.PHONY: run-server
run-server: target/example-server target/example-certs
	target/example-certs
	target/example-server -tokens "alice:wibble" -domains "alice.example.com"

.PHONY: run-all-tests
run-all-tests: run-client-tests run-curl-tests run-grpcurl-tests

# ========================================================================================
# Custom gRPC client
# ========================================================================================

.PHONY: run-client
run-client: target/example-client
	@echo "===> Expect success ..."
	target/example-client -token wibble -msg "G'day"

.PHONY: run-client-alice
run-client-alice: target/example-client
	@echo "===> Expect success ..."
	target/example-client -alice -msg "Tea?"

.PHONY: run-client-bob
run-client-bob: target/example-client
	@echo "===> Expect failure ..."
	-target/example-client -bob -msg "Coffee?"

.PHONY: run-client-tests
run-client-tests: run-client run-client-alice run-client-bob

# ========================================================================================
# Plain HTTP client: curl
# ========================================================================================

.local/bin/jq:
	mkdir -p .local/bin
	curl -s -L -o .local/bin/jq ${JQ_URL}
	chmod +x .local/bin/jq

.PHONY: run-curl
run-curl: .local/bin/jq
	@echo "===> Expect success ..."
	curl --silent --show-error --fail \
		--cacert target/ca.crt \
		-H 'Content-Type: application/json' \
		-H 'Authorization: Bearer wibble' \
		-d '{"message": "hello"}' \
		https://localhost:8443/v1/example/echo | .local/bin/jq '.'

.PHONY: run-curl-alice
run-curl-alice: .local/bin/jq
	@echo "===> Expect success ..."
	curl --silent --show-error --fail \
		--cacert target/ca.crt \
		--cert target/alice.crt \
		--key target/alice.key \
		-H 'Content-Type: application/json' \
		-d '{"message": "Wine?"}' \
		https://localhost:8443/v1/example/echo | .local/bin/jq '.'

.PHONY: run-curl-bob
run-curl-bob:
	@echo "===> Expect failure ..."
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
# ========================================================================================

.local/bin/grpcurl:
	mkdir -p .local/bin
	curl -s -L -o .local/grpcurl.tar.gz ${GRPCURL_URL}
	tar -xzf .local/grpcurl.tar.gz -C .local/bin
	chmod +x .local/bin/grpcurl
	rm .local/grpcurl.tar.gz

.PHONY: run-grpcurl
run-grpcurl: .local/bin/grpcurl
	@echo "===> Expect success ..."
	.local/bin/grpcurl -servername server.example.com \
		-cacert target/ca.crt \
		-H 'authorization: bearer wibble' \
		-d '{"message":"hola"}' \
		localhost:8000 example.service.Example/Echo

.PHONY: run-grpcurl-alice
run-grpcurl-alice: .local/bin/grpcurl
	@echo "===> Expect success ..."
	.local/bin/grpcurl -servername server.example.com \
		-cacert target/ca.crt \
		-cert target/alice.crt \
		-key target/alice.key \
		-d '{"message":"Gin?"}' \
		localhost:8000 example.service.Example/Echo

.PHONY: run-grpcurl-bob
run-grpcurl-bob: .local/bin/grpcurl
	@echo "===> Expect failure ..."
	-.local/bin/grpcurl -servername server.example.com \
		-cacert target/ca.crt \
		-cert target/bob.crt \
		-key target/bob.key \
		-d '{"message":"Vodka?"}' \
		localhost:8000 example.service.Example/Echo

.PHONY: run-grpcurl-tests
run-grpcurl-tests: run-grpcurl run-grpcurl-alice run-grpcurl-bob
