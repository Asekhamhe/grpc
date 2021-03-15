// Client implementation for OrderManagement

package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/asekhamhe/grpc/ordermanagement/client/ecommerce"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
)

const address = "localhost:50051"

type wrappedStream struct {
	grpc.ClientStream
}

// Client-side Unary Interceptor
func orderUnaryClientInterceptor(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Preprocessor phase has access to the RPC request prior to sending it out to the server
	log.Println("Method : " + method)

	// Invoking the remote method
	err := invoker(ctx, method, req, res, cc, opts...)

	// Postprocessor phase
	log.Println(res)

	return err
}

//=======================================================

// Client-side Stream Interceptor
func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Println("======= [Client Interceptor] ", method)
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return newWrappedStream(s), nil
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("======== [Client Stream Interceptor]"+"Receive a message (Type :%T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("======= [Client Stream Interceptor] "+"Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}

func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}

//======================================================= main function

func main() {
	// set up connection to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(orderUnaryClientInterceptor),
		grpc.WithStreamInterceptor(clientStreamInterceptor))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrderManagementClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// ============================== Unary ===================

	// calling gRPC remote GetOrder method
	res, _ := c.GetOrder(ctx, &wrappers.StringValue{Value: "103"})
	log.Println("GetOrder response -> : ", res)

	// calling gRPC remote SearchOrders method
	sst, _ := c.SearchOrders(ctx, &wrappers.StringValue{Value: "Google"})

	for {
		so, err := sst.Recv()
		if err == io.EOF {
			break
		}
		// handle other possible errors
		log.Print("Search Result : ", so)
	}

	// ============================== Streaming ===================

	// // Update Orders : Client streaming scenario

	updOrder1 := pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Google Pixel Book"}, Destination: "Mountain View, CA", Price: 1100.00}
	updOrder2 := pb.Order{Id: "103", Items: []string{"Apple Watch S4", "Mac Book Pro", "iPad Pro"}, Destination: "San Jose, CA", Price: 2800.00}
	updOrder3 := pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub", "iPad Mini"}, Destination: "Mountain View, CA", Price: 2200.00}

	// calling gRPC remote updateOrders method
	updSt, err := c.UpdateOrders(ctx)
	if err != nil {
		log.Fatalf("%v.UpdateOrders(_) = _, %v", c, err)
	}

	// updating order 1
	if err := updSt.Send(&updOrder1); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updSt, updOrder1, err)
	}

	// updating order 2
	if err := updSt.Send(&updOrder2); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updSt, updOrder2, err)
	}

	// updating order 3
	if err := updSt.Send(&updOrder3); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updSt, updOrder3, err)
	}

	updateRes, err := updSt.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", updSt, err, nil)
	}
	log.Printf("Update Orders Res: %s", updateRes)

	// ================== Bidirectional Streaming ===================

	strProcOrd, err := c.ProcessOrders(ctx)
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", c, err)
	}
	if err := strProcOrd.Send(&wrappers.StringValue{Value: "102"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", c, "102", err)
	}
	if err := strProcOrd.Send(&wrappers.StringValue{Value: "103"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", c, "103", err)
	}
	if err := strProcOrd.Send(&wrappers.StringValue{Value: "104"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", c, "104", err)
	}

	ch := make(chan struct{})
	go asyncRPC(strProcOrd, ch)
	time.Sleep(time.Millisecond * 1000)

	if err := strProcOrd.Send(&wrappers.StringValue{Value: "101"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", c, "101", err)
	}
	if err := strProcOrd.CloseSend(); err != nil {
		log.Fatal(err)
	}
	<-ch
}

func asyncRPC(streamProcOrd pb.OrderManagement_ProcessOrdersClient, c chan struct{}) {
	for {
		comShipment, err := streamProcOrd.Recv()
		if err == io.EOF {
			break
		}
		log.Print("Combined shipment : ", comShipment.OrdersList)
	}

	<-c

}
