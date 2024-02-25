package mandelbrot

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/png"
	"log"
	"math"
	"math/cmplx"
	"sync"

	"go-grpc-mandlebrot/pkg/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

const maxIterations = 1000

type MandelbrotServer struct {
	proto.UnimplementedMandelbrotServer
}

func NewMandelbrotServer() *MandelbrotServer {
	return &MandelbrotServer{}
}

func (ms MandelbrotServer) GetImage(_ context.Context, emt *emptypb.Empty) (*proto.Image, error) {
	var imgBuffer bytes.Buffer

	if err := png.Encode(&imgBuffer, ms.generateMandelbrot(1500, 750)); err != nil {
		log.Fatalf("error while encoding image: %v", err)
		return nil, err
	}

	return &proto.Image{ImageContent: imgBuffer.Bytes()}, nil
}

func (MandelbrotServer) generateMandelbrot(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(
		img,
		img.Bounds(),
		&image.Uniform{C: color.White},
		image.Point{},
		draw.Src,
	)

	ratio := width / height

	var wg sync.WaitGroup

	for px := range width {
		wg.Add(1)
		go func(px int) {
			defer wg.Done()
		yLabel:
			for py := range height {

				x := ((float64(2*px) / float64(width)) - 1) * float64(ratio)
				y := (float64(2*py) / float64(height)) - 1

				var z complex128
				c := complex(x, y)

				for n := range maxIterations {
					z = z*z + c

					if cmplx.Abs(z) > 2 {

						// ugly
						logZn := cmplx.Log(z*z) / 2
						nu := cmplx.Log(logZn/cmplx.Log(2)) / cmplx.Log(2)

						color1 := palette.Plan9[int(math.Min(float64(255), float64(n)))]

						color2 := palette.Plan9[int(math.Min(float64(255), float64(n+1)))]

						nRatio := (real(nu) + imag(nu)) - (math.Floor(real(nu) + imag(nu)))

						color := InterpolateColors(color1.(color.RGBA), color2.(color.RGBA), nRatio)
						img.Set(px, py, color)

						continue yLabel
					}
				}
			}

		}(px)
	}

	wg.Wait()

	return img
}

func InterpolateColors(color1, color2 color.RGBA, ratio float64) color.RGBA {
	if ratio <= 0 {
		return color1
	} else if ratio >= 1 {
		return color2
	}

	// Interpolate each color component separately
	red := uint8(float64(color1.R) + ratio*(float64(color2.R)-float64(color1.R)))
	green := uint8(float64(color1.G) + ratio*(float64(color2.G)-float64(color1.G)))
	blue := uint8(float64(color1.B) + ratio*(float64(color2.B)-float64(color1.B)))
	alpha := uint8(float64(color1.A) + ratio*(float64(color2.A)-float64(color1.A)))

	return color.RGBA{R: red, G: green, B: blue, A: alpha}
}
