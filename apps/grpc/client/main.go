package main

import (
	"context"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"

	pb "grpc-gateway-example/pb"
)

func main() {
	log.Println("Staring gRPC client")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// called at program end
	defer closeConnection(cc)

	c := pb.NewYourServiceClient(cc)

	callEcho(err, c)
	callEchoServerStream(err, c)
	callEchoClientStream(err, c)
	callBiDiStream(err, c)
}

func callBiDiStream(e error, client pb.YourServiceClient) {
	stream, err := client.EchoBiDiStream(context.Background())

	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
	}

	waitc := make(chan struct{})

	go func() {

		req := []*pb.StringMessage{
			{
				Value: "val1",
			}, {
				Value: "val2",
			}, {
				Value: "val3",
			},
		}

		for _, v := range req {
			log.Printf("Client streaming: %v", v)
			err := stream.Send(v)

			if err != nil {
				log.Fatalf("Failed to stream to server: %v", err)
			}

			time.Sleep(1000 * time.Millisecond)
		}

		_ = stream.CloseSend()
	}()

	go func() {
		for {
			response, e := stream.Recv()

			if e == io.EOF {
				close(waitc)
				break
			}

			if e != nil {
				log.Fatalf("Failed to recieve: %v", e)
			}

			log.Printf("Received :%v", response)
		}
	}()

	<-waitc
}

func callEchoClientStream(e error, client pb.YourServiceClient) {
	stream, e := client.EchoClientStream(context.Background())

	if e != nil {
		log.Panicf("Failed to open client stream to server: %v", e)
	}

	for i := 0; i < 5; i++ {
		send := stream.Send(&pb.StringMessage{Value: "Value"})

		log.Printf("Send result: %v", send)
	}

	res, e := stream.CloseAndRecv()

	if e != nil {
		log.Panicf("Failed to recieve server response: %v", e)
	}

	log.Printf("Server response: %v", res.GetValue())
}

func callEchoServerStream(e error, client pb.YourServiceClient) {
	stream, e := client.EchoServerStream(context.Background(), &pb.StringMessage{Value: "Value for server to stream"})

	if e != nil {
		log.Fatalf("Failed to read blog: %v", e)
	}

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			// end of stream
			log.Print("Reached EOF")
			break
		}

		if err != nil {
			log.Fatalf("Error while reading stream: %v", err)
		}

		log.Printf("Response from EchoServerStream: %v", msg)
	}
}

func callEcho(e error, client pb.YourServiceClient) {
	res, e := client.Echo(context.Background(), &pb.StringMessage{Value: "Some value"})
	if e != nil {
		log.Panicf("Failed to get echo: %v", e)
	}
	value := res.GetValue()
	log.Printf("Got echo: %s", value)
}

func closeConnection(conn *grpc.ClientConn) {
	e := conn.Close()

	if e != nil {
		log.Fatalf("Failed to close client connection!")
	}

	log.Print("Closed client")
}
