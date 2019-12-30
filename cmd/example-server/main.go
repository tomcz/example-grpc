package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/tomcz/example-grpc/api"
	"github.com/tomcz/example-grpc/server"
)

func main() {
	grpcPort := flag.Int("grpc", 8000, "gRPC listener port")
	httpPort := flag.Int("http", 8080, "HTTP listener port")
	flag.Parse()

	// Fatal logging prevents defer from firing, so wrap the
	// service configuration & startup in a realMain function.
	if err := realMain(*grpcPort, *httpPort); err != nil {
		log.Fatalln("application failed with", err)
	}
	log.Println("application stopped")
}

func realMain(grpcPort, httpPort int) error {
	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)
	defer cancel()

	impl := server.NewExampleServer()

	grpcSrv := newGRPCServer(impl, grpcPort)

	httpSrv, err := newHTTPServer(ctx, impl, httpPort)
	if err != nil {
		return err
	}

	group.Go(func() error { return grpcSrv.ListenAndServe() })
	group.Go(func() error { return httpSrv.ListenAndServe() })

	go waitForShutdown(func() {
		grpcSrv.GracefulStop()
		httpSrv.GracefulStop()
		cancel()
	})
	return group.Wait()
}

func waitForShutdown(callback func()) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("shutting down")
	callback()
}

type apiServer interface {
	ListenAndServe() error
	GracefulStop()
}

type grpcServer struct {
	server *grpc.Server
	port   int
}

func newGRPCServer(impl api.ExampleServer, port int) apiServer {
	srv := grpc.NewServer()
	api.RegisterExampleServer(srv, impl)
	return &grpcServer{
		server: srv,
		port:   port,
	}
}

func (s *grpcServer) ListenAndServe() error {
	log.Println("staring gRPC server on port", s.port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	defer lis.Close()
	return s.server.Serve(lis)
}

func (s *grpcServer) GracefulStop() {
	s.server.GracefulStop()
}

type httpServer struct {
	server *http.Server
	port   int
}

func newHTTPServer(ctx context.Context, impl api.ExampleServer, port int) (apiServer, error) {
	// use least-surprising JSON output options
	marshaller := &runtime.JSONPb{OrigName: true, EmitDefaults: true}
	// yes, we are matching all incoming input as JSON, but see note below
	httpMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaller))
	err := api.RegisterExampleHandlerServer(ctx, httpMux, impl)
	if err != nil {
		return nil, err
	}
	// NOTE: grpc-gateway does not play nice with anything other than JSON request bodies but
	// it does not check that the Content-Type is actually JSON, so let's enforce that a bit.
	handler := handlers.ContentTypeHandler(httpMux, "application/json")
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	return &httpServer{
		server: &srv,
		port:   port,
	}, nil
}

func (s *httpServer) ListenAndServe() error {
	log.Println("staring HTTP server on port", s.port)
	return s.server.ListenAndServe()
}

func (s *httpServer) GracefulStop() {
	s.server.Shutdown(context.Background())
}
