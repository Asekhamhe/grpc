package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "github.com/asekhamhe/grpc/productinfo/client/ecommerce"
)

const address = "localhost:50051"

func main() {
	// Set up connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewProductInfoClient(conn)

	// Contact the server and print out its response.
	name := "Apple iPhone 11"
	description := "Meet Apple iPhone 11. All new dual-camera system with Ultra Wide and Night mode."
	price := float32(699.00)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.AddProduct(ctx, &pb.Product{Name: name, Description: description, Price: price})
	if err != nil {
		log.Fatalf("could not add product: %v", err)
	}
	log.Printf("Product ID: %s added successfully", res.Value)

	p, err := c.GetProduct(ctx, &pb.ProductID{Value: res.Value})
	if err != nil {
		log.Fatalf("could not get product: %v", err)
	}
	log.Printf("Product: %v", p.String())

}
