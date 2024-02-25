package main

import (
	"go-grpc-mandlebrot/internal/config"
	"go-grpc-mandlebrot/internal/proto"
	"go-grpc-mandlebrot/internal/server"
	"go-grpc-mandlebrot/pkg/mandelbrot"
	"go-grpc-mandlebrot/pkg/signal"
	"log"

	"google.golang.org/grpc"
)

func main() {

	cfgReader := config.NewViperConfig("../config.yaml")
	log.Println("Configuration successfully loaded")

	network, err := cfgReader.GetFromGRPC("network")
	if err != nil {
		log.Fatal(err)
	}

	address, err := cfgReader.GetFromGRPC("address")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := server.NewGRPCServer(network, address)

	mandelbrotImgServer := mandelbrot.NewMandelbrotServer()

	proto.RegisterMandelbrotServer(grpcServer.GetServer().(*grpc.Server), mandelbrotImgServer)
	log.Printf("gRPC server listen on server %s", address)

	if err := grpcServer.Serve(); err != nil {
		log.Fatalf("Unable to serve gRPC server: %v", err)
	}

	go signal.ListenSignals()

	defer func() {
		grpcServer.Stop()
	}()
}
