// OrderManagement implementation

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"

	pb "github.com/asekhamhe/grpc/ordermanagement/server/ecommerce"
	"google.golang.org/grpc"
)

const port = ":50051"

var orderMap = make(map[string]pb.Order)

// server is used to implement pb.OrderManagementServer
type server struct {
	orderMap map[string]*pb.Order
	pb.UnimplementedOrderManagementServer
}

// GetOrder server method for client to consume
func (s *server) GetOrder(ctx context.Context, id *wrapper.StringValue) (*pb.Order, error) {
	// service implementation
	ord, _ := orderMap[id.Value]
	return &ord, nil
}

// SearchOrders for streaming query
func (s *server) SearchOrder(q *wrappers.StringValue, st pb.OrderManagement_SearchOrderServer) error {

	for k, o := range orderMap {
		log.Print(k, o)
		for _, i := range o.Items {
			log.Print(i)
			if strings.Contains(i, q.Value) {
				// send the matching orders in a stream
				err := st.Send(&o)
				if err != nil {
					return fmt.Errorf("error sending message to stream : ", err)
				}
				log.Print("Matching Order Found : " + k)
				break

			}
		}
	}
	return nil

}

func main() {
	initSampleData()

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

func initSampleData() {
	orderMap["102"] = pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	orderMap["103"] = pb.Order{
		Id:          "103",
		Items:       []string{"Apple Watch S4"},
		Description: "",
		Price:       400.00,
		Destination: "San Jose, CA",
	}
	orderMap["104"] = pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	orderMap["105"] = pb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	orderMap["106"] = pb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}
