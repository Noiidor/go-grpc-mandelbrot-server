package mandelbrot

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"math/cmplx"
	"math/rand"
	"sync"

	pb "go-grpc-mandlebrot-server/internal/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	maxIters = 1000
)

type MandelbrotServer struct {
	pb.UnimplementedMandelbrotServer
}

func NewMandelbrotServer() *MandelbrotServer {
	return &MandelbrotServer{}
}

type SafeMap struct {
	m  map[int]int
	mx sync.Mutex
}

func NewSafeMap(size int) *SafeMap {
	if size == 0 {
		return &SafeMap{
			m: make(map[int]int),
		}
	} else {
		return &SafeMap{
			m: make(map[int]int, size),
		}
	}
}

// }

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

	ratio := width / height

	//itersForPixel := make(map[Pixel]int, width*height)
	itersForPixel := make([][]float64, width)
	for i := range itersForPixel {
		itersForPixel[i] = make([]float64, height)
	}

	histogram := NewSafeMap(maxIters)

	var wg sync.WaitGroup

	for px := range width {
		wg.Add(1)
		go func(px int) {
			defer wg.Done()
			for py := range height {

				x := ((float64((2 * px)) / float64(width)) - 1) * float64(ratio)
				y := ((float64((2 * py)) / float64(height)) - 1)

				iters := iteratePoint(x, y)

				itersForPixel[px][py] = iters

				if iters < float64(maxIters) {
					histogram.mx.Lock()
					histogram.m[int(math.Floor(iters))]++
					histogram.mx.Unlock()
				}
			}

		}(px)
	}

	wg.Wait()

	total := 0
	for _, v := range histogram.m {
		total += v
	}

	hues := make([]float64, maxIters)
	h := 0.0
	for n := range maxIters {
		h += (float64(histogram.m[n]) / float64(total))
		hues[n] = 1 - h
	}
	hues[len(hues)-1] = h

	for x, column := range itersForPixel {
		for y := range column {

			m := itersForPixel[x][y]
			num := 255 - int(255*linearInterpolation(hues[int(math.Floor(m))], hues[int(math.Ceil(m))], m-float64(math.Floor(m))))
			color := color.RGBA{uint8(num), uint8(num), uint8(num), 255}
			img.Set(x, y, color)
		}
	}

	return img
}

func linearInterpolation(a, b, t float64) float64 {
	return a*(1-t) + b*t
}

func createRandPalette(colors int) color.Palette {
	palette := make(color.Palette, colors)

	for n := range colors {
		color := color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255}
		palette[n] = color
	}

	return palette
}

func highestValue(slice [][]int) int {
	max := 0
	for _, column := range slice {
		for _, cell := range column {
			if cell > max {
				max = cell
			}
		}
	}

	return max
}

func iteratePoint(x, y float64) float64 {
	var z complex128
	c := complex(x, y)

	for n := range maxIters {
		z = z*z + c

		if cmplx.Abs(z) > 2 {
			result := float64(n) + 1.0 - math.Log(math.Log2(cmplx.Abs(z)))
			return result
		}
	}
	return 0
}
