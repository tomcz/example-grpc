# example-grpc

Example of an authenticated gRPC & grpc-gateway HTTP server, and various client connection examples.

## Setting up

1. Install Go 1.13 from https://golang.org/
2. Run `make compile` to compile the server & client code.

## Running the server

All the following `make` targets run a gRPC server on port 8000, and a HTTP server on port 8080, both backed by the same ExampleServer instance. Access to both APIs is protected by bearer token authentication, and optionally by mTLS authentication.

`make run-server` uses service decoration for HTTP & gRPC authentication.

`make run-server-mw` uses HTTP & gRPC middleware for authentication.

`make run-server-mtls` enables mTLS authentication for both HTTP & gRPC APIs.

`make run-server-ref` also enables [gRPC server reflection](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md) to make it easier to test & introspect gRPC APIs.

## Running the clients

`make run-client` runs a gRPC client that sends a token-authenticated request to the gRPC endpoint.

`make run-client-alice` runs a gRPC client that sends a mTLS request to the gRPC endpoint using Alice's certificate & key. It will only work with `make run-server-mtls` since it requires the server to process the client certificate.

`make run-client-bob` runs a gRPC client that sends a mTLS request to the gRPC endpoint using Bob's certificate & key. It will only work with `make run-server-mtls` since it requires the server to process the client certificate.

`make run-curl` invokes curl to send a token-authenticated request to the HTTP server.

`make run-curl-alice` invokes curl to send a mTLS request to the HTTP server using Alice's certificate & key. It will only work with `make run-server-mtls` since it requires the server to process the client certificate.

`make run-curl-bob` invokes curl to send a mTLS request to the HTTP server using Bob's certificate & key. It will only work with `make run-server-mtls` since it requires the server to process the client certificate.

`make run-grpcurl` invokes [grpcurl](https://github.com/fullstorydev/grpcurl) to send a token-authenticated request to the gRPC server. This target uses `.proto` source files to understand the RPC schema used for requests & responses.

`make run-grpcurl-ref` invokes [grpcurl](https://github.com/fullstorydev/grpcurl) to send a token-authenticated request to the gRPC server. This target uses [gRPC server reflection](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md) to understand the RPC schema used for requests & responses, and as such will only work with `make run-server-ref`.

`make run-grpcurl-alice` invokes [grpcurl](https://github.com/fullstorydev/grpcurl) to send a mTLS request to the gRPC server using Alice's certificate & key. It will only work with `make run-server-ref` since it requires gRPC server reflection and for the server to process the client certificate.

`make run-grpcurl-bob` invokes [grpcurl](https://github.com/fullstorydev/grpcurl) to send a mTLS request to the gRPC server using Bob's certificate & key. It will only work with `make run-server-ref` since it requires gRPC server reflection and for the server to process the client certificate.

## Compiling service.proto

If you make any changes to the `service.proto` file you will need to update the generated gRPC and gateway code.

1. Install protoc, either from the official repository at https://github.com/protocolbuffers/protobuf/releases,
or by running `brew install protobuf` if you are on osx.
2. Use `go get` to download the following packages:
    ```
   go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
   go get -u github.com/golang/protobuf/protoc-gen-go
   ```
3. Run `make genproto` from the root of this project's directory.
