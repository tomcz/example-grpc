package server

import (
	"context"
	"log"

	"github.com/golang/protobuf/ptypes"

	"github.com/tomcz/example-grpc/api"
)

type server struct{}

// NewExampleServer constructor
func NewExampleServer() api.ExampleServer {
	return &server{}
}

func (s *server) Echo(_ context.Context, in *api.EchoRequest) (*api.EchoResponse, error) {
	log.Println("Request:", in.Message)
	return &api.EchoResponse{
		Message:   in.Message,
		CreatedAt: ptypes.TimestampNow(),
	}, nil
}
