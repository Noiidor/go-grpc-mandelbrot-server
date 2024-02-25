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

// type Pixel struct {
// 	x, y int
// }

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
	itersForPixel := make([][]int, width)
	for i := range itersForPixel {
		itersForPixel[i] = make([]int, height)
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

				if iters < maxIters {
					histogram.mx.Lock()
					if _, ok := histogram.m[iters]; !ok {
						histogram.m[iters] = 0
					}
					histogram.m[iters]++
					histogram.mx.Unlock()
				}

				// var pixelColor color.Color
				// if iters == 0 {
				// 	pixelColor = color.Black
				// } else {
				// 	colorHue := uint8(iters)
				// 	pixelColor = color.RGBA{colorHue, colorHue, colorHue, 255}
				// }

				// img.Set(px, py, pixelColor)
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
		h += float64(histogram.m[n]) / float64(total)
		hues[n] = h
	}

	//highestIteration := highestValue(itersForPixel)

	//make pallete for number of max iterations
	randPalette := createRandPalette(maxIters)

	//NumIterationsPerPixel := make([]int, highestIteration+1)

	// for _, column := range itersForPixel {
	// 	for _, cell := range column {
	// 		NumIterationsPerPixel[cell]++
	// 	}
	// }

	// total := 0
	// for _, v := range NumIterationsPerPixel {
	// 	total += v
	// }
	// for n := range highestIteration {
	// 	total += NumIterationsPerPixel[n]
	// }

	// hue := make([][]float64, width)
	// for i := range hue {
	// 	hue[i] = make([]float64, height)
	// }

	// for x, column := range itersForPixel {
	// 	for y, cell := range column {
	// 		for n := range cell {
	// 			hue[x][y] += float64(NumIterationsPerPixel[n]) / float64(total)
	// 		}
	// 	}
	// }

	// for x, column := range itersForPixel {
	// 	for y := range column {
	// 		hueTotal := int(math.Round(hue[x][y]))
	// 		if hueTotal >= 1 {
	// 			log.Printf("huetotal: %v", hueTotal)
	// 		}
	// 		color := randPalette[hueTotal]
	// 		img.Set(x, y, color)
	// 	}
	// }

	for x, column := range itersForPixel {
		for y := range column {

			m := itersForPixel[x][y]

			color := randPalette[int(math.Round(hues[m]))]
			img.Set(x, y, color)
		}
	}

	// for k, v := range iterationCounts.m {
	// 	for i := range v {

	// 		if _, ok := hue[k]; !ok {
	// 			hue[k] = 0
	// 		}
	// 		added := float64(NumIterationsPerPixel[i]) / float64(total)
	// 		//log.Printf("added value: %v", added)
	// 		hue[k] += added
	// 	}
	// 	colorIndex := 0
	// 	if iColor, ok := hue[k]; ok {
	// 		colorIndex = int(iColor)
	// 	}
	// 	//log.Printf("color index: %v", colorIndex)
	// 	color := randPalette[colorIndex]
	// 	img.Set(k.x, k.y, color)

	// }

	return img
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

func iteratePoint(x, y float64) int {
	var z complex128
	c := complex(x, y)

	for n := range maxIters {
		z = z*z + c

		if cmplx.Abs(z) > 2 {
			return n
		}
	}
	return 0
}
