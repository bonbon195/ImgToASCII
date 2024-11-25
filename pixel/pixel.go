package pixel

import (
	"image/color"
	"math"
)

func GetLuminance(pixelColor color.Color) float64 {
	r, g, b, _ := pixelColor.RGBA()
	rf := float64(r) / 65535.0
	gf := float64(g) / 65535.0
	bf := float64(b) / 65535.0
	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	if max == 0 {
		return 0
	}
	luminance := (max + min) / 2.0
	return math.Floor(luminance * 10)
}

func LimitLuminance(pixelColor color.Color, scale float64) float64 {
	r, g, b, _ := pixelColor.RGBA()
	rf := float64(r) / 65535.0
	gf := float64(g) / 65535.0
	bf := float64(b) / 65535.0
	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	if max == 0 {
		return 0
	}
	luminance := (max + min) / 2.0
	return math.Floor(luminance*scale) / scale
}

func GetLuminanceGrayscale(pixelColor color.Color) int {
	r, _, _, _ := pixelColor.RGBA()
	if r == 0 {
		return 0
	}
	return int(float64(r) / 65535 * 10)
}

func SetLuminanceGrayscale(pixelColor color.Color, scale float64) uint8 {
	r, _, _, _ := pixelColor.RGBA()
	if r == 0 {
		return 0
	}
	return uint8(math.Floor(float64(r)/65535*scale) / scale * 255)
}

func HSVtoRGB(h, s, v float64) (uint8, uint8, uint8) {
	var r, g, b float64

	hi := int(math.Floor(h/60)) % 6
	vMin := ((100 - s) * v) / 100
	a := (v - vMin) * (math.Mod(h, 60) / 60)
	vInc := vMin + a
	vDec := v - a

	switch hi {
	case 0:
		r, g, b = v, vInc, vMin
	case 1:
		r, g, b = vDec, v, vMin
	case 2:
		r, g, b = vMin, v, vInc
	case 3:
		r, g, b = vMin, vDec, v
	case 4:
		r, g, b = vInc, vMin, v
	case 5:
		r, g, b = v, vMin, vDec
	}

	return uint8(r * 255 / 100), uint8(g * 255 / 100), uint8(b * 255 / 100)
}

func LerpColor(c1, c2 color.Color, weight float64) color.NRGBA {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	var r, g, b uint8
	r = uint8(lerp(float64(r1), float64(r2), weight) / 256)
	g = uint8(lerp(float64(g1), float64(g2), weight) / 256)
	b = uint8(lerp(float64(b1), float64(b2), weight) / 256)

	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}
