package httpx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

type service struct {
	server *http.Server
	mtls   bool
	port   int
}

// NewService creates an HTTP service
func NewService(ctx context.Context, impl api.ExampleServer, port int, auth server.TokenAuth, mtls server.AllowList) (server.Service, error) {
	handler, err := httpHandler(ctx, impl)
	if err != nil {
		return nil, err
	}
	handler = authMiddleware(auth, handler)
	if mtls.Enabled() {
		handler = mtlsMiddleware(mtls, handler)
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	if err = mtlsConfig(srv, mtls); err != nil {
		return nil, err
	}
	return &service{
		server: srv,
		mtls:   mtls.Enabled(),
		port:   port,
	}, nil
}

func httpHandler(ctx context.Context, impl api.ExampleServer) (http.Handler, error) {
	// use least-surprising JSON output options
	marshaller := &runtime.JSONPb{OrigName: true, EmitDefaults: true}
	// yes, we are matching all incoming input as JSON, but see note below
	httpMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaller))
	err := api.RegisterExampleHandlerServer(ctx, httpMux, impl)
	if err != nil {
		return nil, fmt.Errorf("grpc-gateway registration failed: %w", err)
	}
	// NOTE: grpc-gateway does not play nice with anything other than JSON request bodies,
	// unless you want to do your own parsing from HttpBody instances, but it does not check
	// that the Content-Type is actually JSON, so let's enforce that a bit.
	return handlers.ContentTypeHandler(httpMux, "application/json"), nil
}

func mtlsConfig(srv *http.Server, mtls server.AllowList) error {
	if !mtls.Enabled() {
		return nil
	}
	caCert, err := os.ReadFile("pki/ca.crt")
	if err != nil {
		return fmt.Errorf("cannot read root CA cert: %w", err)
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return fmt.Errorf("failed to add root CA cert into cert pool")
	}
	cert, err := tls.LoadX509KeyPair("pki/server.crt", "pki/server.key")
	if err != nil {
		return fmt.Errorf("failed to load cert & key files: %w", err)
	}
	// if we wanted to make mTLS mandatory we should
	// set ClientAuth to tls.RequireAndVerifyClientCert
	cfg := &tls.Config{
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    caCertPool,
		Certificates: []tls.Certificate{cert},
	}
	srv.TLSConfig = cfg
	return nil
}

func (s *service) ListenAndServe() error {
	ll := log.WithField("port", s.port)
	var err error
	if s.mtls {
		ll.Info("starting HTTPS server with mTLS")
		// cert & key files provided during mTLS setup
		err = s.server.ListenAndServeTLS("", "")
	} else {
		ll.Info("starting HTTPS server")
		err = s.server.ListenAndServeTLS("pki/server.crt", "pki/server.key")
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *service) GracefulStop() {
	// let's be nice, but not too nice
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	s.server.Shutdown(ctx)
	cancel()
}
