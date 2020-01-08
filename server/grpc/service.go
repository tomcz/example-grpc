package grpc

import (
	"errors"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type service struct {
	server *grpc.Server
	port   int
}

// NewService creates a gRPC service
func NewService(impl api.ExampleServer, port int, allowReflection bool, opts ...grpc.ServerOption) (server.Service, error) {
	tc, err := credentials.NewServerTLSFromFile("pki/server.crt", "pki/server.key")
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS credentials: %w", err)
	}
	opts = append(opts, grpc.Creds(tc))
	srv := grpc.NewServer(opts...)
	api.RegisterExampleServer(srv, impl)
	if allowReflection {
		reflection.Register(srv)
	}
	return &service{
		server: srv,
		port:   port,
	}, nil
}

func (s *service) ListenAndServe() error {
	log.Println("staring gRPC server on port", s.port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	if err := s.server.Serve(lis); err != nil {
		return err
	}
	// Serve will return a nil error when GracefulStop is called
	// but it's good for us to know that it has been stopped
	return errors.New("gRPC server stopped")
}

func (s *service) GracefulStop() {
	s.server.GracefulStop()
}
