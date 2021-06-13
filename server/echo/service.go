package echo

import (
	"context"

	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type plainServer struct{}

// NewExampleServer vanilla server
func NewExampleServer() api.ExampleServer {
	return &plainServer{}
}

func (s *plainServer) Echo(ctx context.Context, in *api.EchoRequest) (*api.EchoResponse, error) {
	log.WithField("user", server.UserName(ctx)).Info(in.Message)
	return &api.EchoResponse{
		Message:   in.Message,
		CreatedAt: ptypes.TimestampNow(),
	}, nil
}
