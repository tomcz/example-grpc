# example-grpc

Example of an authenticated gRPC & grpc-gateway HTTP server, and various client connection examples.

## Setting up

1. Install Go 1.23 from https://golang.org/
2. Run `make compile` to compile the server & client code.

## Running the server

```
make run-server
```

## Running the clients

1. `make run-client` runs a gRPC client that sends a token-authenticated request to the gRPC endpoint.

2. `make run-client-alice` runs a gRPC client that sends a mTLS request to the gRPC endpoint using Alice's certificate & key.

3. `make run-client-bob` runs a gRPC client that sends a mTLS request to the gRPC endpoint using Bob's certificate & key. It will fail since Bob's certificate is not permitted.

4. `make run-curl` invokes curl to send a token-authenticated request to the HTTP server.

5. `make run-curl-alice` invokes curl to send a mTLS request to the HTTP server using Alice's certificate & key.

6. `make run-curl-bob` invokes curl to send a mTLS request to the HTTP server using Bob's certificate & key. It will fail since Bob's certificate is not permitted.

7. `make run-grpcurl` invokes [grpcurl](https://github.com/fullstorydev/grpcurl) to send a token-authenticated request to the gRPC server.

8. `make run-grpcurl-alice` invokes [grpcurl](https://github.com/fullstorydev/grpcurl) to send a mTLS request to the gRPC server using Alice's certificate & key.

9. `make run-grpcurl-bob` invokes [grpcurl](https://github.com/fullstorydev/grpcurl) to send a mTLS request to the gRPC server using Bob's certificate & key. It will fail since Bob's certificate is not permitted.

## Compiling service.proto

Run `make genproto` from the root of this project's directory.
