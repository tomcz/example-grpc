package grpcx

import (
	"crypto/tls"
	"crypto/x509"
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
func NewService(impl api.ExampleServer, port int, auth server.TokenAuth, mtls server.AllowList) (server.Service, error) {
	authFunc := newServerAuthFunc(auth)
	if mtls.Enabled() {
		authFunc = newMTLSAuthFunc(mtls, authFunc)
	}
	grpcOpts := authMiddleware(authFunc)
	tc, err := newTransportCredentials(mtls.Enabled())
	if err != nil {
		return nil, err
	}
	grpcOpts = append(grpcOpts, grpc.Creds(tc))
	srv := grpc.NewServer(grpcOpts...)
	api.RegisterExampleServer(srv, impl)
	reflection.Register(srv) // make it easy to use grpcurl
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
	log.Println("gRPC server stopped")
	return nil
}

func (s *service) GracefulStop() {
	s.server.GracefulStop()
}
