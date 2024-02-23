package main

import (
	"log"
	"net"

	"github.com/Noiidor/go-grpc-mandelbrot/proto"
	"github.com/Noiidor/go-grpc-mandelbrot/services"
	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		log.Fatal("Unable to listen on port 9000")
	}

	grpcServer := grpc.NewServer()

	imgService := services.MandelbrotServer{}

	proto.RegisterMandelbrotServer(grpcServer, imgService)

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("Unable to serve gRPC server")
	}

}
