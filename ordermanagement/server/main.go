// OrderManagement implementation

package main

import (
	"context"
	"log"
	"net"

	wrapper "github.com/golang/protobuf/ptypes/wrappers"

	pb "github.com/asekhamhe/grpc/ordermanagement/server/ecommerce"
	"google.golang.org/grpc"
)

const port = ":50051"

// server is used to implement pb.OrderManagementServer
type server struct {
	orderMap map[string]*pb.Order
	pb.UnimplementedOrderManagementServer
}

func (s *server) GetOrder(ctx context.Context, id *wrapper.StringValue) (*pb.Order, error) {
	// service implementation
	ord := s.orderMap[id.Value]
	return ord, nil
}

func main() {

	conn, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrderManagementServer(s, &server{})
	log.Printf("starting gRPC listener of port " + port)
	if err := s.Serve(conn); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
