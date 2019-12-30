package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
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
		log.Fatalf("application failed - error is: %v\n", err)
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

	shutdown := func() {
		cancel() // we're done
		grpcSrv.GracefulStop()
		httpSrv.GracefulStop()
	}

	group.Go(func() error { return grpcSrv.ListenAndServe() })
	group.Go(func() error { return httpSrv.ListenAndServe() })

	waitForShutdown(group, shutdown)
	return nil
}

func waitForShutdown(group *errgroup.Group, shutdown func()) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		waitForSignal()
		shutdown()
		wg.Done()
	}()
	go func() {
		err := group.Wait()
		log.Println(err)
		shutdown()
		wg.Done()
	}()
	wg.Wait()
}

func waitForSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("shutdown signal")
}
