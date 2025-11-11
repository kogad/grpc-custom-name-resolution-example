package main

import (
	"context"
	"log"
	"time"

	pb "github.com/kogad/custom-name-resolution/proto"
	_ "github.com/kogad/custom-name-resolution/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect using custom resolver
	// The "custom:///myservice" target will be resolved by our custom resolver
	conn, err := grpc.NewClient(
		"custom:///myservice",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewHelloServiceClient(conn)

	// Send requests every 3 seconds to see the resolver switching servers
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	requestNum := 0
	for range ticker.C {
		requestNum++
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		resp, err := client.SayHello(ctx, &pb.HelloRequest{
			Name: "Client",
		})

		if err != nil {
			log.Printf("Request #%d failed: %v", requestNum, err)
		} else {
			log.Printf("Request #%d - Response: %s", requestNum, resp.Message)
		}

		cancel()
	}
}
