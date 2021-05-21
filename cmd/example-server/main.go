package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/tomcz/example-grpc/server"
	"github.com/tomcz/example-grpc/server/echo"
	"github.com/tomcz/example-grpc/server/grpcx"
	"github.com/tomcz/example-grpc/server/httpx"
)

var (
	grpcPort = flag.Int("grpc", 8000, "gRPC listener port")
	httpPort = flag.Int("http", 8080, "HTTP listener port")
	tokens   = flag.String("tokens", "", "valid bearer tokens")
	domains  = flag.String("domains", "", "valid client TLS certificate domains")
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
	auth := server.NewBearerAuth(*tokens)
	mtls := server.NewDomainAllowList(*domains)

	grpcSrv, err := grpcx.NewService(impl, *grpcPort, auth, mtls)
	if err != nil {
		return err
	}
	httpSrv, err := httpx.NewService(ctx, impl, *httpPort, auth, mtls)
	if err != nil {
		return err
	}

	var group errgroup.Group
	group.Go(func() error {
		defer cancel()
		return grpcSrv.ListenAndServe()
	})
	group.Go(func() error {
		defer cancel()
		return httpSrv.ListenAndServe()
	})
	group.Go(func() error {
		defer func() {
			grpcSrv.GracefulStop()
			httpSrv.GracefulStop()
		}()
		signalChan := make(chan os.Signal)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-signalChan:
			log.Println("shutdown received")
			return nil
		case <-ctx.Done():
			return nil
		}
	})
	return group.Wait()
}
