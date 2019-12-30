package echo

import (
	"context"
	"log"

	auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type authServer struct {
	delegate api.ExampleServer
	apiAuth  server.Auth
}

// NewAuthServer constructor
func NewAuthServer(delegate api.ExampleServer, apiAuth server.Auth) api.ExampleServer {
	return &authServer{
		delegate: delegate,
		apiAuth:  apiAuth,
	}
}

func (s *authServer) Echo(ctx context.Context, in *api.EchoRequest) (*api.EchoResponse, error) {
	token, err := auth.AuthFromMD(ctx, s.apiAuth.Scheme())
	if err != nil {
		return nil, err
	}
	username, err := s.apiAuth.Authenticate(token)
	if err != nil {
		log.Printf("%+v\n", err)
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return s.delegate.Echo(server.WithUserName(ctx, username), in)
}
