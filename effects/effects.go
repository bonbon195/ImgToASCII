package effects

import (
	"ascii/pixel"
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
)

func GaussianBlur2D(im image.Image, sigma float64) *image.NRGBA {
	bounds := im.Bounds()
	blurredImage := image.NewNRGBA(bounds)
	kernel := generateGaussianKernel2D(sigma)
	radius := len(kernel) / 2
	w := new(sync.WaitGroup)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		w.Add(1)
		go func(y int) {
			defer w.Done()
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				var newR, newG, newB float64
				for ky := -radius; ky <= radius; ky++ {
					for kx := -radius; kx <= radius; kx++ {
						clampedX := clampToBorders(x+kx, bounds.Min.X, bounds.Max.X-1)
						clampedY := clampToBorders(y+ky, bounds.Min.Y, bounds.Max.Y-1)
						r, g, b, _ := im.At(clampedX, clampedY).RGBA()
						weight := kernel[ky+radius][kx+radius]
						newR += weight * float64(r>>8)
						newG += weight * float64(g>>8)
						newB += weight * float64(b>>8)
					}
				}
				color := color.NRGBA{R: clamp(int(newR)), G: clamp(int(newG)), B: clamp(int(newB)), A: 255}
				blurredImage.Set(x, y, color)
			}
		}(y)
	}
	w.Wait()
	return blurredImage
}

func GaussianBlur(im image.Image, sigma float64) *image.NRGBA {
	return gaussianBlurVertical(gaussianBlurHorizontal(im, sigma), sigma)
}

func GenerateAsciiFiles(im image.Image, asciiTexture []string, addColors bool) error {
	bounds := im.Bounds()
	bordersImage := GaussianDifference(im, 0.5, 6, 120)
	bordersImage = SobelOperatorAngleColored(bordersImage, 1200)
	grayscaleImage := imaging.AdjustSaturation(im, -100)
	art := AsciiBorders(bordersImage, 4)
	w := new(sync.WaitGroup)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 8 {
		w.Add(1)
		go func(y int) {
			defer w.Done()
			for x := bounds.Min.X; x < bounds.Max.X; x += 8 {
				if art[y/8][x/8] != "" {
					continue
				}
				blockWidth := min(8, bounds.Max.X-x)
				pixelColor := grayscaleImage.At(x, y)
				luminance := pixel.GetLuminanceGrayscale(pixelColor)
				// 1-10 -> 0-9 because this is used as index
				if luminance > 0 {
					luminance--
				}
				letter := asciiTexture[luminance]
				art[y/8][x/8] = letter
				if x == bounds.Max.X-blockWidth {
					art[y/8][x/8+1] = "\n"
				}
			}
		}(y)
	}
	w.Wait()
	if addColors {
		art = AsciiAddColors(im, art, 8)
	}

	var result string
	for i := 0; i < len(art); i++ {
		result += strings.Join(art[i], "")
	}

	err := os.WriteFile("ascii-result.html", []byte(generateHTML(result)), 0666)
	if err != nil {
		return errors.New("Couldn't write to html file: " + err.Error())
	}
	return nil
}

func AsciiBorders(im image.Image, threshold int) [][]string {
	bounds := im.Bounds()
	wg := new(sync.WaitGroup)
	art := make([][]string, bounds.Max.Y/8)
	for i := 0; i < len(art); i++ {
		art[i] = make([]string, bounds.Max.X/8+1)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += 8 {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			for x := bounds.Min.X; x < bounds.Max.X; x += 8 {

				// we either have 8x8 or less block
				blockHeight := min(8, bounds.Max.Y-y)
				blockWidth := min(8, bounds.Max.X-x)

				verticalSum := 0
				horizontalSum := 0
				diagonalFrontSum := 0
				diagonalBackSum := 0
				for by := 0; by < blockHeight; by++ {
					for bx := 0; bx < blockWidth; bx++ {
						clr := im.At(x+bx, y+by).(color.NRGBA)

						switch clr {
						case color.NRGBA{R: 0, G: 0, B: 255, A: 255}:
							verticalSum++
						case color.NRGBA{R: 0, G: 255, B: 0, A: 255}:
							diagonalFrontSum++
						case color.NRGBA{R: 255, G: 0, B: 0, A: 255}:
							horizontalSum++
						case color.NRGBA{R: 255, G: 255, B: 0, A: 255}:
							diagonalBackSum++
						}
					}
				}
				letter := determineEdgeLetter(horizontalSum, verticalSum, diagonalFrontSum, diagonalBackSum, threshold)
				if letter == "" {
					continue
				}
				art[y/8][x/8] = letter
				if x == bounds.Max.X-blockWidth {
					art[y/8][x/8+1] = "\n"
				}
			}
		}(y)
	}
	wg.Wait()
	return art
}

func GaussianDifference(im image.Image, sigma, k float64, threshold int) *image.NRGBA {
	im = imaging.AdjustSaturation(im, -100)
	// im = imaging.Grayscale(im)
	bounds := im.Bounds()
	w := new(sync.WaitGroup)
	var blurred *image.NRGBA
	var blurred2 *image.NRGBA
	w.Add(2)
	go func() {
		defer w.Done()
		blurred = GaussianBlur(im, sigma)
	}()
	go func() {
		defer w.Done()
		blurred2 = GaussianBlur(im, k*sigma)
	}()
	w.Wait()
	newImage := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			r1, g1, b1, _ := blurred.NRGBAAt(x, y).RGBA()
			r2, g2, b2, _ := blurred2.NRGBAAt(x, y).RGBA()

			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			var tau float64 = 0.4
			rDiff := int(math.Floor((1+tau)*float64(r1) - tau*float64(r2)))
			gDiff := int(math.Floor((1+tau)*float64(g1) - tau*float64(g2)))
			bDiff := int(math.Floor((1+tau)*float64(b1) - tau*float64(b2)))

			// Apply threshold to produce black and white output
			if rDiff > threshold || gDiff > threshold || bDiff > threshold {
				newImage.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			} else {
				newImage.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
			}

			// rDiff := int(r1 - r2)
			// gDiff := int(g1 - g2)
			// bDiff := int(b1 - b2)

			// newImage.SetNRGBA(x, y, color.NRGBA{R: clamp(rDiff), G: clamp(gDiff), B: clamp(bDiff)})
		}
	}
	return newImage
}

func SobelOperator(im image.Image, threshhold float64) *image.NRGBA {
	bounds := im.Bounds()
	newImage := image.NewNRGBA(bounds)
	radius := len(sobelKernelHorizontal) / 2

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var sumX int
			var sumY int
			for ky := -radius; ky <= radius; ky++ {
				for kx := -radius; kx <= radius; kx++ {
					clampedX := clampToBorders(x+kx, bounds.Min.X, bounds.Max.X)
					clampedY := clampToBorders(y+ky, bounds.Min.Y, bounds.Max.Y)
					r, _, _, _ := im.At(clampedX, clampedY).RGBA()
					sumX += sobelKernelHorizontal[ky+radius][kx+radius] * int(r)
					sumY += sobelKernelVertical[ky+radius][kx+radius] * int(r)

				}
			}
			magnitute := math.Sqrt(float64(sumX*sumX + sumY*sumY))
			if magnitute > threshhold {
				newImage.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			} else {
				newImage.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
			}
		}
	}
	return newImage
}

func SobelOperatorAngleColored(im image.Image, threshold float64) *image.NRGBA {
	bounds := im.Bounds()
	newImage := image.NewNRGBA(bounds)
	radius := len(sobelKernelHorizontal) / 2

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var sumX int
			var sumY int
			for ky := -radius; ky <= radius; ky++ {
				for kx := -radius; kx <= radius; kx++ {
					clampedX := clampToBorders(x+kx, bounds.Min.X, bounds.Max.X)
					clampedY := clampToBorders(y+ky, bounds.Min.Y, bounds.Max.Y)
					r, _, _, _ := im.At(clampedX, clampedY).RGBA()
					sumX += sobelKernelHorizontal[ky+radius][kx+radius] * int(r)
					sumY += sobelKernelVertical[ky+radius][kx+radius] * int(r)

				}
			}
			magnitute := math.Sqrt(float64(sumX*sumX + sumY*sumY))
			if magnitute < threshold {
				newImage.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
				continue
			}

			angle := math.Atan2(float64(sumY), float64(sumX))
			normalizedAngle := (angle + math.Pi) / (2 * math.Pi) // Normalize to 0-1
			normalizedAngle = math.Mod(normalizedAngle*360, 180) // Normalize to 0-180 degrees
			var slopeColor color.NRGBA
			switch {
			case normalizedAngle < 22.5 || normalizedAngle >= 157.5:
				slopeColor = color.NRGBA{R: 0, G: 0, B: 255, A: 255} // Horizontal (_)
			case normalizedAngle >= 112.5 && normalizedAngle < 157.5:
				slopeColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255} // Diagonal should be (/) but it's actually (\)
			case normalizedAngle >= 67.5 && normalizedAngle < 112.5:
				slopeColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255} // Vertical (|)
			case normalizedAngle >= 22.5 && normalizedAngle < 67.5:
				slopeColor = color.NRGBA{R: 255, G: 255, B: 0, A: 255} // Diagonal should be (\) but it's actually (/)
			default:
				slopeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			}
			newImage.Set(x, y, slopeColor)
		}
	}
	return newImage
}

func SobelOperatorColored(im image.Image, threshold float64) *image.NRGBA {
	bounds := im.Bounds()
	newImage := image.NewNRGBA(bounds)
	radius := len(sobelKernelHorizontal) / 2

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var sumX int
			var sumY int
			for ky := -radius; ky <= radius; ky++ {
				for kx := -radius; kx <= radius; kx++ {
					clampedX := clampToBorders(x+kx, bounds.Min.X, bounds.Max.X)
					clampedY := clampToBorders(y+ky, bounds.Min.Y, bounds.Max.Y)
					r, _, _, _ := im.At(clampedX, clampedY).RGBA()
					sumX += sobelKernelHorizontal[ky+radius][kx+radius] * int(r)
					sumY += sobelKernelVertical[ky+radius][kx+radius] * int(r)

				}
			}
			magnitute := math.Sqrt(float64(sumX*sumX + sumY*sumY))
			if magnitute < threshold {
				newImage.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
				continue
			}

			angle := math.Atan2(float64(sumY), float64(sumX))
			normalizedAngle := (angle + math.Pi) / (2 * math.Pi) // Normalize to 0-1
			r, g, b := pixel.HSVtoRGB(normalizedAngle*360, 100, 100)
			newImage.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return newImage
}

func AsciiAddColors(im image.Image, art [][]string, scale int) [][]string {
	bounds := im.Bounds()
	resized := imaging.Resize(im, bounds.Max.X/scale, bounds.Max.Y/scale, imaging.NearestNeighbor)
	resized = imaging.Resize(im, bounds.Max.X, bounds.Max.Y, imaging.NearestNeighbor)
	w := &sync.WaitGroup{}
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 8 {
		w.Add(1)
		go func(y int) {
			defer w.Done()
			for x := bounds.Min.X; x < bounds.Max.X; x += 8 {
				if art[y/scale][x/scale] == "\n" {
					continue
				}
				var r, g, b uint32
				r, g, b, _ = resized.At(x, y).RGBA()
				r, g, b = r>>8, g>>8, b>>8
				art[y/scale][x/scale] = fmt.Sprintf("<span style='color: rgb(%d, %d, %d);'>%s</span>", r, g, b, art[y/scale][x/scale])
			}
		}(y)
	}
	w.Wait()
	return art
}

func ResizeLerp(im image.Image, width, height int) *image.NRGBA {
	bounds := im.Bounds()
	newImage := image.NewNRGBA(image.Rect(0, 0, width, height))
	widthScale := float64(bounds.Max.X) / float64(width)
	heightScale := float64(bounds.Max.Y) / float64(height)
	log.Println(widthScale, heightScale)
	s := 0
	s2 := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			x1 := clampToBorders(int(math.Floor(float64(x)*widthScale)), 0, bounds.Dx()-1)
			y1 := clampToBorders(int(math.Floor(float64(y)*heightScale)), 0, bounds.Dy()-1)
			x2 := clampToBorders(int(math.Ceil(float64(x)*widthScale)), 0, bounds.Dx()-1)
			y2 := clampToBorders(int(math.Ceil(float64(y)*heightScale)), 0, bounds.Dy()-1)

			xWeight := widthScale*float64(x) - float64(x1)
			yWeight := heightScale*float64(y) - float64(y1)

			pix1 := im.At(x1, y1)
			pix2 := im.At(x2, y1)
			pix3 := im.At(x1, y2)
			pix4 := im.At(x2, y2)

			newColor := pixel.LerpColor(
				pixel.LerpColor(pix1, pix2, xWeight),
				pixel.LerpColor(pix3, pix4, xWeight),
				yWeight,
			)
			newImage.SetNRGBA(x, y, newColor)
		}
	}
	log.Println(s, s2)
	return newImage
}
