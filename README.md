# example-grpc

## Compiling service.proto

1. Install protoc, either from the official repository at https://github.com/protocolbuffers/protobuf/releases,
or by running `brew install protobuf` if you are on osx.
2. Use `go get` to download the following packages:
    ```
   go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
   go get -u github.com/golang/protobuf/protoc-gen-go
   ```
3. Run `make genproto` from the root of this project's directory.
