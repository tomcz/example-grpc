package main

import (
	"context"
	"flag"
	"log"

	rpc "google.golang.org/grpc"

	"github.com/tomcz/example-grpc/server/auth"
	"github.com/tomcz/example-grpc/server/echo"
	"github.com/tomcz/example-grpc/server/grpc"
	"github.com/tomcz/example-grpc/server/http"
)

var (
	grpcPort   = flag.Int("grpc", 8000, "gRPC listener port")
	httpPort   = flag.Int("http", 8080, "HTTP listener port")
	tokens     = flag.String("tokens", "alice:wibble,bob:letmein", "valid bearer tokens")
	middleware = flag.Bool("middleware", false, "apply authentication middleware")
	reflection = flag.Bool("reflection", false, "enable server reflection API")
	mtls       = flag.Bool("mtls", false, "enable mTLS between server & clients")
)

func main() {
	flag.Parse()
	// Fatal logging prevents defer from firing, so wrap the
	// service configuration & startup in a realMain function.
	if err := realMain(); err != nil {
		log.Fatalf("application failed - error is: %v\n", err)
	}
	log.Println("application stopped")
}

func realMain() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	impl := echo.NewExampleServer()
	authn := auth.NewBearerAuth(*tokens)

	var httpMiddleware []http.Middleware
	var grpcMiddleware []rpc.ServerOption
	if *middleware {
		log.Println("using HTTP & gRPC middleware for authentication")
		httpMiddleware = http.AuthMiddleware(authn)
		grpcMiddleware = grpc.AuthMiddleware(authn)
	} else {
		log.Println("using service decorator for authentication")
		impl = auth.NewAuthDecorator(impl, authn)
	}

	grpcSrv, err := grpc.NewService(impl, *grpcPort, *reflection, *mtls, grpcMiddleware...)
	if err != nil {
		return err
	}
	httpSrv, err := http.NewService(ctx, impl, *httpPort, *mtls, httpMiddleware...)
	if err != nil {
		return err
	}

	shutdown := func() {
		grpcSrv.GracefulStop()
		httpSrv.GracefulStop()
		cancel()
	}
	runAndWaitForExit(shutdown, grpcSrv.ListenAndServe, httpSrv.ListenAndServe)
	return nil
}
