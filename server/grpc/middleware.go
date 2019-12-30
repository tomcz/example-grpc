package grpc

import (
	auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"

	"github.com/tomcz/example-grpc/server"
)

// AuthMiddleware optional authentication middleware
func AuthMiddleware(authn server.Auth) []grpc.ServerOption {
	authFunc := server.NewAuthFunc(authn)
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(authFunc)),
		grpc.StreamInterceptor(auth.StreamServerInterceptor(authFunc)),
	}
}
