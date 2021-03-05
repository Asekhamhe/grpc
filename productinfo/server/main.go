// ProductInfo bussiness logic implementation for gRPC

package main

import (
	"context"
	"log"
	"net"

	pb "github.com/asekhamhe/grpc/productinfo/server/ecommerce"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const port = ":50051"

// server is used to implement pb.ProductInfoServer
type server struct {
	productMap map[string]*pb.Product
	// Embed the unimplemented server
	pb.UnimplementedProductInfoServer
}

// AddProduct implements ecommerce.AddProduct
func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	out, err := uuid.NewUUID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while generating product id", err)
	}
	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in
	return &pb.ProductID{Value: in.Id}, status.New(codes.OK, "").Err()

}

// GetProduct implements ecommerce.GetProduct
func (s *server) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {
	v, ok := s.productMap[in.Value]
	if ok {
		return v, status.New(codes.OK, "").Err()
	}
	return nil, status.Errorf(codes.NotFound, "product does not exist.", in.Value)
}

func main() {
	conn, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterProductInfoServer(s, &server{})
	log.Printf("starting gRPC listener on port " + port)
	if err := s.Serve(conn); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
