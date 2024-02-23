package main

import (
	"bytes"
	"context"
	"image/png"
	"log"
	"os"

	"github.com/Noiidor/go-grpc-client/client"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	client, err := client.NewClient()
	if err != nil {
		log.Fatalf("Cant create gRPC client: %v", err)
	}

	img, err := client.GetImage(context.Background(), new(emptypb.Empty))
	if err != nil {
		log.Fatalf("Cannot get image: %v", err)
	}

	imgBytes := img.ImageContent

	buffer := bytes.NewBuffer(imgBytes)

	decodedImg, err := png.Decode(buffer)
	if err != nil {
		log.Fatalf("Error during decoding image buffer: %v", err)
	}

	imgFile, err := os.Create("decoded.png")
	if err != nil {
		log.Fatalf("Cannot create image file: %v", err)
	}
	defer imgFile.Close()

	if err = png.Encode(imgFile, decodedImg); err != nil {
		log.Fatalf("Cannot encode image bytes into file: %v", err)
	}

	log.Print("Image received successfully")

}
