package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/tomcz/example-grpc/server/echo"
	"github.com/tomcz/example-grpc/server/grpc"
	"github.com/tomcz/example-grpc/server/http"
)

var (
	grpcPort = flag.Int("grpc", 8000, "gRPC listener port")
	httpPort = flag.Int("http", 8080, "HTTP listener port")
	tokens   = flag.String("tokens", "wibble,letmein", "valid bearer tokens")
)

func main() {
	flag.Parse()
	// Fatal logging prevents defer from firing, so wrap the
	// service configuration & startup in a realMain function.
	if err := realMain(); err != nil {
		log.Fatalln("application failed with", err)
	}
	log.Println("application stopped")
}

func realMain() error {
	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)
	defer cancel()

	impl := echo.NewExampleServer(strings.Split(*tokens, ","))
	grpcSrv := grpc.NewService(impl, *grpcPort)
	httpSrv, err := http.NewService(ctx, impl, *httpPort)
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
