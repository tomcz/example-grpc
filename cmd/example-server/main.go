package main

import (
	"context"
	"flag"
	"log"

	"github.com/tomcz/example-grpc/server/bearer"
	"github.com/tomcz/example-grpc/server/echo"
	"github.com/tomcz/example-grpc/server/grpcx"
	"github.com/tomcz/example-grpc/server/httpx"
)

var (
	grpcPort = flag.Int("grpc", 8000, "gRPC listener port")
	httpPort = flag.Int("http", 8080, "HTTP listener port")
	tokens   = flag.String("tokens", "alice:wibble,bob:letmein", "valid bearer tokens")
	mtls     = flag.Bool("mtls", false, "enable mTLS between server & clients")
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
	auth := bearer.NewBearerAuth(*tokens)

	grpcSrv, err := grpcx.NewService(impl, *grpcPort, auth, *mtls)
	if err != nil {
		return err
	}
	httpSrv, err := httpx.NewService(ctx, impl, *httpPort, auth, *mtls)
	if err != nil {
		return err
	}

	shutdown := func() {
		cancel()
		grpcSrv.GracefulStop()
		httpSrv.GracefulStop()
	}
	runAndWaitForExit(
		shutdown,
		grpcSrv.ListenAndServe,
		httpSrv.ListenAndServe,
	)
	return nil
}
