package main

import (
	"context" // Use "golang.org/x/net/context" for Golang version <= 1.6
	"flag"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	pb "grpc-gateway-example/pb"
)

var (
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:50051", "gRPC server endpoint")
)

type server struct {
}

func (server) EchoBiDiStream(stream pb.YourService_EchoBiDiStreamServer) error {
	log.Println("EchoBiDiStream called")

	//isLastMessage := false
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			//isLastMessage = true
			return nil
		}

		if err != nil && err != io.EOF {
			log.Fatalf("Error while reading client stream")
			return err
		}

		value := req.GetValue()
		log.Printf("Value from client: %v", value)

		sendErr := stream.Send(&pb.StringMessage{Value: value})

		if sendErr != nil {
			log.Fatalf("Error while writing to client")
			return sendErr
		}
	}

	return nil
}

func (server) EchoClientStream(stream pb.YourService_EchoClientStreamServer) error {
	log.Println("EchoClientStream called")

	var lastClientMessage string

	for {
		req, err := stream.Recv()

		if req != nil {
			log.Printf("Message from client: %v", req)
			lastClientMessage = req.GetValue()
		}

		if err == io.EOF {
			// we can respond whenever we want
			log.Println("Finished reading from client")
			return stream.SendAndClose(&pb.StringMessage{Value: lastClientMessage})
		}

		if err != nil {
			log.Fatalf("Error reading from client stream: %v", err)
		}
	}
}

func (server) EchoServerStream(req *pb.StringMessage, stream pb.YourService_EchoServerStreamServer) error {
	log.Println("EchoServerStream called")

	value := req.GetValue()

	for i := 0; i < 5; i++ {
		send := stream.Send(&pb.StringMessage{
			Value: value,
		})

		log.Printf("Send result: %v", send)
		time.Sleep(1000 * time.Millisecond)
	}

	return nil
}

func (server) Echo(ctx context.Context, req *pb.StringMessage) (*pb.StringMessage, error) {
	log.Println("Echo called")

	value := req.GetValue()
	log.Printf("Echo called with: %v", value)

	return &pb.StringMessage{
		Value: value,
	}, nil
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// register gRPC server

	lis, err := net.Listen("tcp", *grpcServerEndpoint)

	if err != nil {
		log.Fatalf("Failed to listen : %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterYourServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	go func() {
		log.Println("Starting gRPC server")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to server: %v", err)
		}
	}()

	// wait for Ctrl + C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until signal is received
	<-ch

	log.Println("Stopping the server")
	s.Stop()

	log.Println("Closing the listener")
	_ = lis.Close()

	return nil
}

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
