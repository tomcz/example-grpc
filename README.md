# example-grpc

Example of an authenticated gRPC & grpc-gateway server, and various client connection examples.

## Setting up

1. Install Go 1.13 from https://golang.org/
2. Run `make compile` to compile the server & client code.

## Running the server

All the following `make` targets run a gRPC server on port 8000, and a HTTP server on port 8080, both backed by the same ExampleServer instance. Access to both APIs is protected by bearer token authentication.

`make run-server` uses service decoration for HTTP & gRPC authentication.

`make run-server-mw` uses HTTP & gRPC middleware for authentication.

`make run-server-ref` also enables [gRPC server reflection](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md) to make it easier to test & introspect gRPC APIs.

## Running the clients

All the following `make` targets send an authenticated request to the server.

`make run-client` runs a gRPC client that sends a request to the gRPC server.

`make run-curl` invokes curl to send a request to the HTTP server.

`make run-grpcurl` uses [grpcurl](https://github.com/fullstorydev/grpcurl) to send a request to the gRPC server. This target uses `.proto` source files to understand the RPC schema used for requests & responses.

`make run-grpcurl-ref` uses [grpcurl](https://github.com/fullstorydev/grpcurl) to send a request to the gRPC server. This target uses [gRPC server reflection](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md) to understand the RPC schema used for requests & responses, and as such will only work with `make run-server-ref`.

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
