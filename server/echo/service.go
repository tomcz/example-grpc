package echo

import (
	"context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type plainServer struct {
	api.UnimplementedExampleServer
}

// NewExampleServer vanilla server
func NewExampleServer() api.ExampleServer {
	return &plainServer{}
}

func (s *plainServer) Echo(ctx context.Context, in *api.EchoRequest) (*api.EchoResponse, error) {
	log.WithField("user", server.UserName(ctx)).Info(in.Message)
	return &api.EchoResponse{
		Message:   in.Message,
		CreatedAt: timestamppb.Now(),
	}, nil
}
