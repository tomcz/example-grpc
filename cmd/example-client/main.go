package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"

	"github.com/tomcz/example-grpc/api"
)

func main() {
	addr := flag.String("addr", "localhost:8000", "server address")
	msg := flag.String("msg", "G'day", "message to send")
	flag.Parse()

	// Fatal logging prevents defer from firing, so wrap the
	// service configuration & startup in a realMain function.
	if err := realMain(*addr, *msg); err != nil {
		log.Fatalln("application failed with", err)
	}
	log.Println("application stopped")
}

func realMain(addr, msg string) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := api.NewExampleClient(conn)
	res, err := client.Echo(context.Background(), &api.EchoRequest{Message: msg})
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
