package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/kogad/custom-name-resolution/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedHelloServiceServer
	addr string
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	message := fmt.Sprintf("Hello %s from %s", req.Name, s.addr)
	log.Printf("Received request from %s, responding with: %s", req.Name, message)
	return &pb.HelloResponse{Message: message}, nil
}

func main() {
	port := flag.Int("port", 8888, "The server port")
	flag.Parse()

	addr := fmt.Sprintf("localhost:%d", *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, &server{addr: addr})

	log.Printf("Server listening on %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
