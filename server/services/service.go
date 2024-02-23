package services

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math/cmplx"
	"sync"

	pb "github.com/Noiidor/go-grpc-mandelbrot/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MandelbrotServer struct {
	pb.UnimplementedMandelbrotServer
}

func (MandelbrotServer) GetImage(ctx context.Context, emt *emptypb.Empty) (*pb.Image, error) {

	var imgBuffer bytes.Buffer

	img := generateMandelbrot(1500, 750)

	err := png.Encode(&imgBuffer, img)
	if err != nil {
		log.Fatalf("Error while encoding image: %v", err)
		return nil, err
	}

	return &pb.Image{ImageContent: imgBuffer.Bytes()}, nil
}

func generateMandelbrot(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{0, 0}, draw.Src)

	yStep := 1.0 / float64(img.Bounds().Dy())
	xStep := 1.0 / float64(img.Bounds().Dx())

	xOffset := float64(img.Bounds().Dx()) / 1.5
	yOffset := float64(img.Bounds().Dy()) / 2

	scaleModifier := float64(img.Bounds().Dy()) / 2

	var wg sync.WaitGroup

	for i := float64(-1.0); i < 1; i += yStep {
		wg.Add(1)
		go func(iCaptured float64) {
			defer wg.Done()

		xLabel:
			for j := float64(-2); j < 1; j += xStep {
				var z complex128
				c := complex(j, iCaptured)

				iterCount := 0

				x, y := transformIntoImgCoords(j, iCaptured, xOffset, yOffset, scaleModifier)

				for k := 0; k <= 40; k++ {
					z = z*z + c
					iterCount++
					if cmplx.Abs(z) > 2 {

						color := color.RGBA{uint8(255 - 5*iterCount), uint8(255 - 5*iterCount), uint8(255 - 5*iterCount), 255}

						img.Set(x, y, color)

						continue xLabel
					}

				}

				img.Set(x, y, color.Black)
			}
		}(i)
	}

	wg.Wait()

	return img
}

func transformIntoImgCoords(xFloat, yFloat, xOffset, yOffset, scaleModifier float64) (x, y int) {

	xFloat *= scaleModifier
	yFloat *= scaleModifier

	xFloat += xOffset
	yFloat += yOffset

	return int(xFloat), int(yFloat)
}
