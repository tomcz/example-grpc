package echo

import (
	"context"

	auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type authDecorator struct {
	delegate api.ExampleServer
	authFunc auth.AuthFunc
}

// NewAuthDecorator applies authentication to every request
// and only delegates successfully authenticated requests.
func NewAuthDecorator(delegate api.ExampleServer, authn server.Auth) api.ExampleServer {
	return &authDecorator{
		delegate: delegate,
		authFunc: server.NewAuthFunc(authn),
	}
}

func (s *authDecorator) Echo(ctx context.Context, in *api.EchoRequest) (*api.EchoResponse, error) {
	authCtx, err := s.authFunc(ctx)
	if err != nil {
		return nil, err
	}
	return s.delegate.Echo(authCtx, in)
}
