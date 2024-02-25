package server

import (
	"google.golang.org/grpc"
	"log"
	"net"
)

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
}

func NewGRPCServer(network, address string) Server {
	gc := &GRPCServer{}

	gc.init(network, address)

	return gc
}

func (gc *GRPCServer) init(network, address string) {
	var err error

	gc.listener, err = net.Listen(network, address)
	if err != nil {
		log.Fatal("Unable to listen on port 9000")
	}

	gc.server = grpc.NewServer()
}

func (gc *GRPCServer) GetServer() interface{} {
	return gc.server
}

func (gc *GRPCServer) Serve() error {
	return gc.server.Serve(gc.listener)
}

func (gc *GRPCServer) Stop() {
	gc.server.Stop()
}
