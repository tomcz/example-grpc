package grpc

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type service struct {
	server *grpc.Server
	port   int
}

// NewService creates a gRPC service
func NewService(impl api.ExampleServer, port int, opts ...grpc.ServerOption) server.Service {
	srv := grpc.NewServer(opts...)
	api.RegisterExampleServer(srv, impl)
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
	defer lis.Close()
	return s.server.Serve(lis)
}

func (s *service) GracefulStop() {
	s.server.GracefulStop()
}
