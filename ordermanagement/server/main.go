// OrderManagement implementation

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

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

type wrappedStream struct {
	grpc.ServerStream
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

// Server :: Unary Interceptor
func orderUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Pre-processing logic
	// Gets info about the current RPC call by examining the args passed in
	log.Println("======= [Server Interceptor] ", info.FullMethod)
	log.Printf("Pre Processing Message: %s", req)

	// Invoking the handler to complete the normal execution of a unary RPC.
	m, err := handler(ctx, req)

	// Post processing logic
	log.Printf("Post Processing Message : %s", m)

	return m, err

}

// Server :: Stream Interceptor
func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("==== [Server Stream Interceptor Wrapper] "+"Receive a message (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("===== =[Server Stream Interceptor Wrapper] "+"Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.SendMsg(m)
}

// newWrappedStream
func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s}
}

func orderServerStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("====== [Server Stream Interceptor] ", info.FullMethod)
	err := handler(srv, newWrappedStream(ss))
	if err != nil {
		log.Printf("rpc failed with error %v", err)
	}
	return err
}

func main() {
	initSampleData()

	conn, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(orderUnaryServerInterceptor),
		grpc.StreamInterceptor(orderServerStreamInterceptor),
	)
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
