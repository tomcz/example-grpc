GOPATH = $(shell go env GOPATH)

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

.PHONY: genproto
genproto:
	protoc -Iapi/ -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:api \
		api/service.proto
	protoc -Iapi/ -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:api \
		api/service.proto

.PHONY: compile
compile: compile-server compile-client

.PHONY: compile-server
compile-server: target
	go build -o target/example-server ./cmd/example-server/...

.PHONY: compile-client
compile-client: target
	go build -o target/example-client ./cmd/example-client/...

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
	./target/example-client -bob -msg "Coffee?" || true

.PHONY: run-client-tests
run-client-tests: run-client run-client-alice run-client-bob

# ========================================================================================
# Plain HTTP client: curl
# ========================================================================================

.PHONY: run-curl
run-curl:
	curl --silent --show-error --fail --cacert pki/ca.crt \
		-H 'Content-Type: application/json' \
		-H 'Authorization: Bearer wibble' \
		-d '{"message": "hello"}' https://localhost:8080/v1/example/echo | jq '.'

.PHONY: run-curl-alice
run-curl-alice:
	curl --silent --show-error --fail --cacert pki/ca.crt \
		--key pki/alice.key --cert pki/alice.crt \
		-H 'Content-Type: application/json' \
		-d '{"message": "Wine?"}' https://localhost:8080/v1/example/echo | jq '.'

.PHONY: run-curl-bob
run-curl-bob:
	curl --silent --show-error --fail --cacert pki/ca.crt \
 		--key pki/bob.key --cert pki/bob.crt \
		-H 'Content-Type: application/json' \
		-d '{"message": "Whiskey?"}' https://localhost:8080/v1/example/echo || true

.PHONY: run-curl-tests
run-curl-tests: run-curl run-curl-alice run-curl-bob

# ========================================================================================
# Third-party gRPC client: grpcurl
# ========================================================================================

.PHONY: run-grpcurl
run-grpcurl:
	grpcurl -cacert pki/ca.crt -servername server.example.com \
		-d '{"message":"hola"}' -H 'authorization: bearer wibble' \
		localhost:8000 example.service.Example/Echo

.PHONY: run-grpcurl-alice
run-grpcurl-alice:
	grpcurl -cacert pki/ca.crt -servername server.example.com \
		-cert pki/alice.crt -key pki/alice.key \
		-d '{"message":"Gin?"}' \
		localhost:8000 example.service.Example/Echo

.PHONY: run-grpcurl-bob
run-grpcurl-bob:
	grpcurl -cacert pki/ca.crt -servername server.example.com \
		-cert pki/bob.crt -key pki/bob.key \
		-d '{"message":"Vodka?"}' \
		localhost:8000 example.service.Example/Echo || true

.PHONY: run-grpcurl-tests
run-grpcurl-tests: run-grpcurl run-grpcurl-alice run-grpcurl-bob
