package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
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
	middleware = flag.Bool("middleware", false, "apply authentiation middleware")
)

func main() {
	flag.Parse()
	// Fatal logging prevents defer from firing, so wrap the
	// service configuration & startup in a realMain function.
	if err := realMain(); err != nil {
		log.Fatalf("application failed with %+v\n", err)
	}
	log.Println("application stopped")
}

func realMain() error {
	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)
	defer cancel()

	impl := echo.NewExampleServer()
	authn := auth.NewBearerAuth(*tokens)

	var httpMiddleware []http.Middleware
	var grpcMiddleware []rpc.ServerOption
	if *middleware {
		log.Println("using HTTP & gRPC middleware for authentication")
		httpMiddleware = []http.Middleware{http.AuthMiddleware(authn)}
		grpcMiddleware = grpc.AuthMiddleware(authn)
	} else {
		log.Println("using service wrapper for authentication")
		impl = echo.NewAuthServer(impl, authn)
	}

	grpcSrv := grpc.NewService(impl, *grpcPort, grpcMiddleware...)
	httpSrv, err := http.NewService(ctx, impl, *httpPort, httpMiddleware...)
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
