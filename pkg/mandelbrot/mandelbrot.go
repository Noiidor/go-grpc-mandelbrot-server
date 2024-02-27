package mandelbrot

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math/cmplx"
	"math/rand"
	"sync"

	pb "go-grpc-mandlebrot-server/internal/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	maxIters         = 4000
	thresholdIters   = 4000
	regionPercentage = 1
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

type ColorRegion struct {
	startIter  int
	startColor color.RGBA
}

func (MandelbrotServer) GetImage(ctx context.Context, emt *emptypb.Empty) (*pb.Image, error) {

	var imgBuffer bytes.Buffer

	img := generateMandelbrot(2000, 1000)

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

	itersForPixel := make([][]int, width)
	for i := range itersForPixel {
		itersForPixel[i] = make([]int, height)
	}

	histogram := &SafeMap{m: make(map[int]int, maxIters)} // мапа для гистограммы: кол-во итераций-счетчик

	var wg sync.WaitGroup

	for px := range width {
		wg.Add(1)
		go func(px int) {
			defer wg.Done()
			for py := range height {

				x := ((float64((2 * px)) / float64(width)) - 1) * float64(ratio)
				y := ((float64((2 * py)) / float64(height)) - 1)

				iters := iteratePoint(x, y) // главный алгоритм, возвращает кол-во итераций для ухода в бесконечность на заданных координатах

				itersForPixel[px][py] = iters // массив(типо мапа) координаты(как ключ)-итерация(значение)

				if iters < maxIters {
					histogram.mx.Lock()
					histogram.m[iters]++
					histogram.mx.Unlock()
				}
			}

		}(px)
	}

	wg.Wait()

	// total := 0
	// for _, v := range histogram.m {
	// 	total += v
	// }
	colorPlottedPixels(itersForPixel, img)

	return img
}

func colorPlottedPixels(pixelsIterations [][]int, img *image.RGBA) {

	itersToColor := 0
	if maxIters < thresholdIters {
		itersToColor = maxIters
	} else {
		itersToColor = thresholdIters
	}

	itersPerRegion := numOfPercentage(itersToColor, regionPercentage)
	numOfRegions := itersToColor / itersPerRegion

	bands := make([]ColorRegion, numOfRegions)

	for i := range numOfRegions {
		startIter := itersToColor - (itersPerRegion * (i + 1))
		if i == numOfRegions-1 {
			if startIter != 0 {
				startIter = 0
			}
		}

		region := ColorRegion{
			startIter:  startIter,
			startColor: getRandomRGBAColor(),
		}
		bands[i] = region
	}

	for x, pixelsColumn := range pixelsIterations { // цикл по всем уже вычисленным пикселям
		for y, n := range pixelsColumn {

			var currentColor color.Color

			if n == maxIters {
				currentColor = color.Black
			} else {
				for i, region := range bands {
					if n >= region.startIter && n <= region.startIter+itersPerRegion {

						ratio := ratioBetweenNums(region.startIter, region.startIter+itersPerRegion, n)
						var endColor color.RGBA
						if i-1 < 0 {
							endColor = color.RGBA{0, 0, 0, 0}
						} else {
							endColor = bands[i-1].startColor
						}
						currentColor = lerpColor(region.startColor, endColor, ratio)
						break

					} else {
						continue
					}
				}
				if currentColor == nil {
					currentColor = color.Black
				}
			}

			img.Set(x, y, currentColor)
		}
	}

}

func ratioBetweenNums(a, b, x int) float64 {
	return 1.0 - (float64(b)-float64(x))/(float64(b)-float64(a))
}

func getRandomRGBAColor() color.RGBA {
	return color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255}
}

func numOfPercentage(numFrom, percentage int) int {
	return (percentage * numFrom) / 100
}

func percentageOfNum(numFrom, number int) int {
	return (number * 100) / numFrom
}

func lerp(a, b, t float64) float64 {
	return (a * (1.0 - t)) + (b * t)
}

func lerpColor(a, b color.RGBA, t float64) color.Color {

	if t == 0 {
		return a
	}
	if t == 1.0 {
		return b
	}

	// Простите
	resultColor := color.RGBA{
		uint8(lerp(float64(a.R), float64(b.R), t)),
		uint8(lerp(float64(a.G), float64(b.R), t)),
		uint8(lerp(float64(a.B), float64(b.R), t)),
		uint8(lerp(float64(a.A), float64(b.A), t))}

	return resultColor
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
	return maxIters
}
