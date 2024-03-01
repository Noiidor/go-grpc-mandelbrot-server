package service

import (
	"bytes"
	"context"
	pb "go-grpc-mandlebrot-server/internal/proto"
	"go-grpc-mandlebrot-server/pkg/mandelbrot"
	"image/png"
	"log"
)

type MandelbrotServer struct {
	pb.UnimplementedMandelbrotServer
}

func NewMandelbrotServer() *MandelbrotServer {
	return &MandelbrotServer{}
}

func (MandelbrotServer) GetImage(ctx context.Context, settings *pb.MandelbrotSettings) (*pb.Image, error) {

	var imgBuffer bytes.Buffer

	img := mandelbrot.GenerateMandelbrot(int(settings.Width), int(settings.Height), int(settings.Zoom), float64(settings.CenterX), float64(settings.CenterY))

	err := png.Encode(&imgBuffer, img)
	if err != nil {
		log.Fatalf("Error while encoding image: %v", err)
		return nil, err
	}

	return &pb.Image{ImageContent: imgBuffer.Bytes()}, nil
}
