package effects

import (
	"image"
	"image/color"
	"math"
	"sync"
)

func clamp(value int) uint8 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}

func generateGaussianKernel2D(sigma float64) [][]float64 {
	size := int(math.Ceil(sigma * 3))
	if size%2 == 0 {
		size += 1
	}
	radius := size / 2
	kernel := make([][]float64, size)
	for i := 0; i < size; i++ {
		kernel[i] = make([]float64, size)
	}
	var sum float64 = 0
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			value := gaussianKernelFormula2D(float64(x), float64(y), sigma)
			kernel[y+radius][x+radius] = value
			sum += value
		}
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			kernel[y][x] /= sum
		}
	}
	return kernel
}

func generateGaussianKernel(sigma float64) []float64 {
	size := int(math.Ceil(sigma * 3))
	if size%2 == 0 {
		size++
	}
	var sum float64 = 0
	kernel := make([]float64, size)
	for i := 0; i < size; i++ {
		value := gaussianKernelFormula(float64(i), sigma)
		kernel[i] = value
		sum += value
	}
	for x := 0; x < size; x++ {
		kernel[x] /= sum
	}
	return kernel
}

func gaussianBlurVertical(im image.Image, sigma float64) *image.NRGBA {
	bounds := im.Bounds()
	blurredImage := image.NewNRGBA(bounds)
	kernel := generateGaussianKernel(sigma)
	radius := len(kernel) / 2
	w := new(sync.WaitGroup)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		w.Add(1)
		go func(x int) {
			defer w.Done()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				var newR, newG, newB float64
				for ky := -radius; ky <= radius; ky++ {
					clampedY := clampToBorders(y+ky, bounds.Min.Y, bounds.Max.Y-1)
					r, g, b, _ := im.At(x, clampedY).RGBA()
					weight := kernel[ky+radius]
					newR += weight * float64(r>>8)
					newG += weight * float64(g>>8)
					newB += weight * float64(b>>8)
				}
				// color := color.NRGBA{R: clamp(int(newR)), G: clamp(int(newG)), B: clamp(int(newB)), A: 255}
				color := color.NRGBA{R: clamp(int(newR)), G: clamp(int(newG)), B: clamp(int(newB)), A: 255}
				blurredImage.Set(x, y, color)
			}
		}(x)
	}
	w.Wait()
	return blurredImage
}

func gaussianBlurHorizontal(im image.Image, sigma float64) *image.NRGBA {
	bounds := im.Bounds()
	blurredImage := image.NewNRGBA(bounds)
	kernel := generateGaussianKernel(sigma)
	radius := len(kernel) / 2
	w := new(sync.WaitGroup)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		w.Add(1)
		go func(y int) {
			defer w.Done()
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				var newR, newG, newB float64
				for kx := -radius; kx <= radius; kx++ {
					clampedX := clampToBorders(x+kx, bounds.Min.X, bounds.Max.X-1)
					r, g, b, _ := im.At(clampedX, y).RGBA()
					weight := kernel[kx+radius]
					newR += weight * float64(r>>8)
					newG += weight * float64(g>>8)
					newB += weight * float64(b>>8)
				}
				color := color.NRGBA{R: clamp(int(newR)), G: clamp(int(newG)), B: clamp(int(newB)), A: 255}
				blurredImage.Set(x, y, color)
			}
		}(y)
	}
	w.Wait()
	return blurredImage
}

func GaussTest(im image.Image, amount int) image.Image {
	for i := 0; i < amount; i++ {
		im = GaussTestV(GaussTestH(im))
	}
	return im
}

func GaussTestV(im image.Image) *image.NRGBA {
	bounds := im.Bounds()
	blurredImage := image.NewNRGBA(bounds)
	kernel := gaussKernelTest()
	w := new(sync.WaitGroup)
	radius := len(kernel) / 2

	for x := 0; x < bounds.Max.X; x++ {
		w.Add(1)
		go func(x int) {
			defer w.Done()
			for y := 0; y < bounds.Max.Y; y++ {
				var newR, newB, newG float64
				for ky := -radius; ky <= radius; ky++ {
					clampedY := clampToBorders(y+ky, bounds.Min.Y, bounds.Max.Y-1)
					r, g, b, _ := im.At(x, clampedY).RGBA()
					weight := kernel[ky+radius]

					newR += weight * float64(r>>8)
					newG += weight * float64(g>>8)
					newB += weight * float64(b>>8)
				}
				color := color.NRGBA{R: clamp(int(newR)), G: clamp(int(newG)), B: clamp(int(newB)), A: 255}
				blurredImage.Set(x, y, color)
			}
		}(x)
	}
	w.Wait()

	return blurredImage
}

func GaussTestH(im image.Image) *image.NRGBA {
	bounds := im.Bounds()
	blurredImage := image.NewNRGBA(bounds)
	kernel := gaussKernelTest()
	w := new(sync.WaitGroup)
	radius := len(kernel) / 2
	for y := 0; y < bounds.Max.Y; y++ {
		w.Add(1)
		go func(y int) {
			defer w.Done()
			for x := 0; x < bounds.Max.X; x++ {
				var newR, newB, newG float64
				for kx := -radius; kx <= radius; kx++ {
					clampedX := clampToBorders(x+kx, bounds.Min.X, bounds.Max.X-1)
					r, g, b, _ := im.At(clampedX, y).RGBA()
					weight := kernel[kx+radius]

					newR += weight * float64(r>>8)
					newG += weight * float64(g>>8)
					newB += weight * float64(b>>8)
				}

				color := color.NRGBA{R: clamp(int(newR)), G: clamp(int(newG)), B: clamp(int(newB)), A: 255}
				blurredImage.Set(x, y, color)
			}
		}(y)
	}
	w.Wait()

	return blurredImage
}

func gaussKernelTest() []float64 {
	size := 5
	sigma := 1.0
	var sum float64 = 0
	kernel := make([]float64, size)
	for i := 0; i < size; i++ {
		value := gaussianKernelFormula(float64(i), sigma)
		kernel[i] = value
		sum += value
	}
	for x := 0; x < size; x++ {
		kernel[x] /= sum
	}
	return kernel
}

func gaussianKernelFormula2D(x, y, sigma float64) float64 {
	return 1 / (2 * math.Pi * sigma * sigma) * math.Exp(-(x*x+y*y)/(2*sigma*sigma))
}

func gaussianKernelFormula(x, sigma float64) float64 {
	return 1 / math.Sqrt(2*math.Pi*sigma*sigma) * math.Exp(-(x*x)/(2*sigma*sigma))
}

func determineEdgeLetter(horizontalSum, verticalSum, diagonalFrontSum, diagonalBackSum, threshold int) string {
	max := max(verticalSum, max(horizontalSum, max(diagonalBackSum, diagonalFrontSum)))
	if max < threshold {
		return ""
	}
	switch max {
	case verticalSum:
		return "|"
	case horizontalSum:
		return "_"
	case diagonalFrontSum:
		return "\\"
	case diagonalBackSum:
		return "/"
	}
	return ""
}

func generateHTML(asciiArt string) string {
	return `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=0.4">
		<title>RaiMei</title>
	</head>
	<body >
		<pre>` + asciiArt + `</pre>
	</body>
	<style>
		body{
			background: black;
			margin: 0px;
			padding: 0px;
			display: flex;
			justify-content: center;
			height: 100vh;
  			width: 100vw;
		}

		pre {
			background: inherit;
			color: white;
			font-family: monospace;
			letter-spacing: 0.5em;
			font-size: 0.4em;

		}

		pre .height-scaling {
			letter-spacing: 0.2vh;
			font-size: 0.4vh;
		}

	</style>
	</html>`
}

func clampToBorders(coord, boundMin, boundMax int) int {
	if coord < boundMin {
		return boundMin
	}
	if coord > boundMax {
		return boundMax
	}
	return coord
}

var sobelKernelHorizontal = [][]int{
	{-1, 0, 1},
	{-2, 0, 2},
	{-1, 0, 1},
}

var sobelKernelVertical = [][]int{
	{-1, -2, -1},
	{0, 0, 0},
	{1, 2, 1},
}
