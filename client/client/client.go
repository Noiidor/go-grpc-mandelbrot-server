package client

import (
	pb "github.com/Noiidor/go-grpc-client/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient() (pb.MandelbrotClient, error) {
	var options []grpc.DialOption

	options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial("localhost:9000", options...)
	if err != nil {
		return nil, err
	}

	client := pb.NewMandelbrotClient(conn)
	return client, nil
}
