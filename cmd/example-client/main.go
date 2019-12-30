package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"

	"github.com/tomcz/example-grpc/api"
)

func main() {
	// Fatal logging prevents defer from firing, so wrap the
	// service configuration & startup in a realMain function.
	if err := realMain(); err != nil {
		log.Fatalln("application failed with", err)
	}
	log.Println("application stopped")
}

func realMain() error {
	conn, err := grpc.Dial("localhost:8000", grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := api.NewExampleClient(conn)
	res, err := client.Echo(context.Background(), &api.EchoRequest{Message: "G'day"})
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
