package utils

import (
	"image"
	"os"

	"github.com/disintegration/imaging"
)

func OpenFile(filename string) (image.Image, error) {
	f, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	im, err := imaging.Decode(f)
	if err != nil {
		return nil, err
	}
	return im, nil
}
