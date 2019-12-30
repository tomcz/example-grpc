# example-grpc

Example of a gRPC & grpc-gateway client & server.

`make run-server` runs a gRPC server on port 8000, and a HTTP server on port 8080, both backed by the same ExampleServer instance. Access to both APIs is protected by bearer token authentication.

`make run-grpc-client` runs a gRPC client that sends a request to the server, using bearer token authentication.

`make run-http-client` invokes curl to do exactly the same to the HTTP server.

## Setting up

1. Install Go 1.13 from https://golang.org/
2. Run `make compile` to compile the server & client code.

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
