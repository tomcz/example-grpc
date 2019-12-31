package grpc

import (
	"errors"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type service struct {
	server *grpc.Server
	port   int
}

// NewService creates a gRPC service
func NewService(impl api.ExampleServer, port int, allowReflection bool, opts ...grpc.ServerOption) server.Service {
	srv := grpc.NewServer(opts...)
	api.RegisterExampleServer(srv, impl)
	if allowReflection {
		reflection.Register(srv)
	}
	return &service{
		server: srv,
		port:   port,
	}
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
	// server will return a nil error when GracefulStop
	// is called, but we cannot have a silent exit here
	// since we use errgroups, and it requires non-nil
	// errors to cancel the group, otherwise it hangs
	return errors.New("gRPC server stopped")
}

func (s *service) GracefulStop() {
	s.server.GracefulStop()
}
