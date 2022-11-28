package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/tomcz/gotools/errgroup"
	"github.com/tomcz/gotools/quiet"

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
		log.WithError(err).Fatal("application failed")
	}
	log.Info("application stopped")
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

	shutdown := &quiet.Closer{}
	shutdown.AddFunc(grpcSrv.GracefulStop)
	shutdown.AddFunc(httpSrv.GracefulStop)

	group := errgroup.New()
	group.Go(func() error {
		defer cancel()
		return grpcSrv.ListenAndServe()
	})
	group.Go(func() error {
		defer cancel()
		return httpSrv.ListenAndServe()
	})
	group.Go(func() error {
		defer shutdown.CloseAll()
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-signalChan:
			log.Info("shutdown received")
			return nil
		case <-ctx.Done():
			return nil
		}
	})
	return group.Wait()
}
