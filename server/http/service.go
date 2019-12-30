package http

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

// Middleware allows the creation of handler chains
type Middleware func(http.Handler) http.Handler

type service struct {
	server *http.Server
	port   int
}

// NewService creates a HTTP service
func NewService(ctx context.Context, impl api.ExampleServer, port int, middleware ...Middleware) (server.Service, error) {
	// use least-surprising JSON output options
	marshaller := &runtime.JSONPb{OrigName: true, EmitDefaults: true}
	// yes, we are matching all incoming input as JSON, but see note below
	httpMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaller))
	err := api.RegisterExampleHandlerServer(ctx, httpMux, impl)
	if err != nil {
		return nil, errors.Wrap(err, "grpc-gateway registration failed")
	}
	// NOTE: grpc-gateway does not play nice with anything other than JSON request bodies but
	// it does not check that the Content-Type is actually JSON, so let's enforce that a bit.
	handler := handlers.ContentTypeHandler(httpMux, "application/json")
	// apply any additional middleware
	for _, mw := range middleware {
		handler = mw(handler)
	}
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	return &service{
		server: &srv,
		port:   port,
	}, nil
}

func (s *service) ListenAndServe() error {
	log.Println("staring HTTP server on port", s.port)
	return s.server.ListenAndServe()
}

func (s *service) GracefulStop() {
	s.server.Shutdown(context.Background())
}
