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
	maxIters = 2000
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

	histogram := NewSafeMap(maxIters) // мапа для гистограммы: кол-во итераций-счетчик

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

	for x, column := range itersForPixel { // цикл по всем уже вычисленным пикселям
		for y := range column {

			var finalColor color.Color
			n := itersForPixel[x][y]          // итерация на текущем пикселе
			percent := percentageOfMaxIter(n) // Каким процентом от макс.предела итераций текущая итерация является
			switch {                          // секции градиентов палитры
			case n == 0:
				finalColor = color.Black // если итераций на пикселе 0 - значит пиксель внутри множества и не раскрашивается(черный)
			case 5 > percent:
				finalColor = color.RGBA{uint8(n * 10), 25, 25, 255}
			case 10 > percent:
				finalColor = color.RGBA{uint8(n), 50, 50, 255}
			case 20 > percent:
				finalColor = color.RGBA{uint8(n - 200 + 50), 50, 50, 255} // рандомная формула для рассчета градиента
			case 40 > percent:
				finalColor = color.RGBA{255, uint8(n - 400 + 50), 0, 255} // попробуй посчитать вручную в калькуляторе, станет понятнее что происходит и откуда такая формула
			case 65 > percent:
				finalColor = color.RGBA{255, uint8(n - 600 + 50), 0, 255} // RGBA имеет 4 значения, последнее - Alpha, всегда 255, RGB - красный, зеленый, синий
			case 85 > percent:
				finalColor = color.RGBA{255, uint8(n - 800 + 50), 0, 255}
			default:
				finalColor = color.RGBA{255, 255, 0, 255}
			}
			img.Set(x, y, finalColor) // отрисовка пикселя
		}
	}

	return img
}

func percentageOfMaxIter(iter int) int {
	return (iter * 100) / maxIters
}

func lerp(a, b, t float64) float64 {
	return (a * (1.0 - t)) + (b * t)
}

func lerpColor(a, b color.Color, t float64) color.Color {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()

	// Простите
	resultColor := color.RGBA{
		uint8(lerp(float64(r1), float64(r2), t)),
		uint8(lerp(float64(g1), float64(g2), t)),
		uint8(lerp(float64(b1), float64(b2), t)),
		uint8(lerp(float64(a1), float64(a2), t))}

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
	return 0
}
