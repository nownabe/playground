package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"buf.build/gen/go/nownabe/petapis/grpc/go/pet/v1/petv1grpc"
	petv1 "buf.build/gen/go/nownabe/petapis/protocolbuffers/go/pet/v1"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	listenOn := "127.0.0.1:8090"
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenOn, err)
	}

	server := grpc.NewServer()
	petv1grpc.RegisterPetStoreServiceServer(server, &petStoreServiceServer{})
	log.Println("Listening on", listenOn)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

type petStoreServiceServer struct {
	petv1grpc.UnimplementedPetStoreServiceServer
}

func (s *petStoreServiceServer) PutPet(ctx context.Context, req *petv1.PutPetRequest) (*petv1.PutPetResponse, error) {
	name := req.GetName()
	petType := req.GetPetType()
	log.Println("Got a request to create a", petType, "named", name)

	return &petv1.PutPetResponse{}, nil
}
