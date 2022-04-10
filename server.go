package main

import (
	"log"
	"net"
	"os"

	"sloth-grpc/sql_service"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	listen, err := net.Listen("tcp", "0.0.0.0:"+os.Getenv("PORT"))

	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", os.Getenv("PORT"), err)
	}

	log.Println("Listening @ : " + os.Getenv("PORT"))

	grpcserver := grpc.NewServer()

	cs := sql_service.SQLServiceServer{}

	sql_service.RegisterSQLServicesServer(grpcserver, &cs)

	err = grpcserver.Serve(listen)

	if err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}
}
