// OrderManagement implementation

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"

	pb "github.com/asekhamhe/grpc/ordermanagement/server/ecommerce"
	"google.golang.org/grpc"
)

const (
	port           = ":50051"
	orderBatchSize = 3
)

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
func (s *server) SearchOrders(q *wrappers.StringValue, st pb.OrderManagement_SearchOrdersServer) error {

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

// updateOrders implementation for streaming RPC
func (s *server) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {
	ostr := "Updated Order IDs : "
	for {
		o, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&wrappers.StringValue{Value: "Orders processed " + ostr})
		}
		if err != nil {
			return nil
		}

		// Update order
		orderMap[o.Id] = *o

		log.Printf("Order ID : %s - %s", o.Id, "Updated")
		ostr += o.Id + ", "
	}
}

// Bi-directional Streaming RPC for ProcessOrders
func (s *server) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {
	batchMaker := 1
	var comShipMap = make(map[string]pb.CombineShipment)
	for {
		oID, err := stream.Recv()
		log.Printf("Reading Proc order : %s", oID)
		if err == io.EOF {
			// Client has sent all the messages
			// Send remaining shipments
			log.Printf("EOF : %s ", oID)
			for _, shipment := range comShipMap {
				if err := stream.Send(&shipment); err != nil {
					return err
				}

			}
			return nil
		}
		if err != nil {
			log.Println(err)
			return err
		}
		dest := orderMap[oID.GetValue()].Destination
		shipment, ok := comShipMap[dest]

		// destination found
		if ok {
			ord := orderMap[oID.GetValue()]
			shipment.OrdersList = append(shipment.OrdersList, &ord)
			comShipMap[dest] = shipment
		} else {
			comShip := pb.CombineShipment{Id: " cmb - " + (orderMap[oID.GetValue()].Destination), Status: "Processed!"}
			ord := orderMap[oID.GetValue()]
			comShip.OrdersList = append(shipment.OrdersList, &ord)
			comShipMap[dest] = comShip
			log.Print(len(comShip.OrdersList), comShip.GetId())
		}

		if batchMaker == orderBatchSize {
			for _, comb := range comShipMap {
				log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrdersList))
				if err := stream.Send(&comb); err != nil {
					return err
				}
			}
			batchMaker = 0
			comShipMap = make(map[string]pb.CombineShipment)
		} else {
			batchMaker++
		}
	}
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
