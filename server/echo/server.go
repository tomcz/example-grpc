package echo

import (
	"context"
	"log"

	"github.com/golang/protobuf/ptypes"
	auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tomcz/example-grpc/api"
)

type server struct {
	validTokens map[string]bool
}

// NewExampleServer constructor
func NewExampleServer(tokens []string) api.ExampleServer {
	tokenMap := make(map[string]bool)
	for _, token := range tokens {
		tokenMap[token] = true
	}
	return &server{
		validTokens: tokenMap,
	}
}

func (s *server) Echo(ctx context.Context, in *api.EchoRequest) (*api.EchoResponse, error) {
	err := s.authenticate(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Request:", in.Message)
	return &api.EchoResponse{
		Message:   in.Message,
		CreatedAt: ptypes.TimestampNow(),
	}, nil
}

func (s *server) authenticate(ctx context.Context) error {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return err
	}
	if !s.validTokens[token] {
		return status.Error(codes.PermissionDenied, "invalid token")
	}
	return nil
}
