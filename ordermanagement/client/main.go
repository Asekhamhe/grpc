// Client implementation for OrderManagement

package main

import (
	"context"
	"log"
	"time"

	pb "github.com/asekhamhe/grpc/ordermanagement/client/ecommerce"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
)

const address = "localhost:50051"

func main() {
	// set up connection to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrderManagementClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.GetOrder(ctx, &wrappers.StringValue{Value: "103"})

	log.Println("GetOrder response -> : ", res)
}
