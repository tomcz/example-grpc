package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
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
func NewService(impl api.ExampleServer, port int, allowReflection bool, allowMtls bool, opts ...grpc.ServerOption) (server.Service, error) {
	tc, err := newTransportCredentials(allowMtls)
	if err != nil {
		return nil, err
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

func newTransportCredentials(allowMtls bool) (credentials.TransportCredentials, error) {
	if allowMtls {
		return newMTLSTransportCredentials()
	}
	return credentials.NewServerTLSFromFile("pki/server.crt", "pki/server.key")
}

func newMTLSTransportCredentials() (credentials.TransportCredentials, error) {
	caCert, err := ioutil.ReadFile("pki/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("cannot read root CA cert: %w", err)
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to add root CA cert into cert pool")
	}
	cert, err := tls.LoadX509KeyPair("pki/server.crt", "pki/server.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load cert & key files: %w", err)
	}
	// if we wanted to make mTLS mandatory we should
	// set ClientAuth to tls.RequireAndVerifyClientCert
	cfg := &tls.Config{
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    caCertPool,
		Certificates: []tls.Certificate{cert},
	}
	cfg.BuildNameToCertificate()
	return credentials.NewTLS(cfg), nil
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
