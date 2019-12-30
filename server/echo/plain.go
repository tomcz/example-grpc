package echo

import (
	"context"
	"log"

	"github.com/golang/protobuf/ptypes"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type plainServer struct{}

// NewPlainServer constructor
func NewPlainServer() api.ExampleServer {
	return &plainServer{}
}

func (s *plainServer) Echo(ctx context.Context, in *api.EchoRequest) (*api.EchoResponse, error) {
	log.Printf("user: %v - request: %v\n", server.UserName(ctx), in.Message)
	return &api.EchoResponse{
		Message:   in.Message,
		CreatedAt: ptypes.TimestampNow(),
	}, nil
}
