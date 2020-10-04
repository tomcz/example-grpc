package httpx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type service struct {
	server *http.Server
	mtls   bool
	port   int
}

// NewService creates a HTTP service
func NewService(ctx context.Context, impl api.ExampleServer, port int, auth server.TokenAuth, mtls server.AllowList) (server.Service, error) {
	// use least-surprising JSON output options
	marshaller := &runtime.JSONPb{OrigName: true, EmitDefaults: true}
	// yes, we are matching all incoming input as JSON, but see note below
	httpMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaller))
	err := api.RegisterExampleHandlerServer(ctx, httpMux, impl)
	if err != nil {
		return nil, fmt.Errorf("grpc-gateway registration failed: %w", err)
	}
	// NOTE: grpc-gateway does not play nice with anything other than JSON request bodies but
	// it does not check that the Content-Type is actually JSON, so let's enforce that a bit.
	handler := handlers.ContentTypeHandler(httpMux, "application/json")
	handler = authMiddleware(auth, handler)
	if mtls.Enabled() {
		handler = mtlsMiddleware(mtls, handler)
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	if mtls.Enabled() {
		cfg, err := mtlsConfig()
		if err != nil {
			return nil, err
		}
		srv.TLSConfig = cfg
	}
	return &service{
		server: srv,
		mtls:   mtls.Enabled(),
		port:   port,
	}, nil
}

func mtlsConfig() (*tls.Config, error) {
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
	return cfg, nil
}

func (s *service) ListenAndServe() error {
	log.Println("staring HTTP server on port", s.port)
	if s.mtls {
		// cert & key files provided during mTLS setup
		return s.server.ListenAndServeTLS("", "")
	}
	return s.server.ListenAndServeTLS("pki/server.crt", "pki/server.key")
}

func (s *service) GracefulStop() {
	s.server.Shutdown(context.Background())
}
