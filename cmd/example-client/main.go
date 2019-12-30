package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/tomcz/example-grpc/api"
)

var (
	token = flag.String("token", "wibble", "authentication token")
	addr  = flag.String("addr", "localhost:8000", "server address")
	msg   = flag.String("msg", "G'day", "message to send")
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
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	md := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", *token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	client := api.NewExampleClient(conn)
	res, err := client.Echo(ctx, &api.EchoRequest{Message: *msg})
	if err != nil {
		return err
	}
	marshaller := &jsonpb.Marshaler{
		EmitDefaults: true,
		OrigName:     true,
		Indent:       "  ",
	}
	txt, err := marshaller.MarshalToString(res)
	if err != nil {
		return err
	}
	fmt.Println(txt)
	return nil
}
