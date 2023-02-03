package main

import (
	"grpc-gamelobby-boilerplate/server/api"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	log.Printf("running server on 9000")

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	var l *server
	if l, err = initServer(); err != nil {
		log.Fatal("failed to init server")
	}
	api.RegisterLobbyServiceServer(s, l)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
