package main

import (
	"ascii/effects"
	"ascii/utils"
	"log"
	"os"
	"strconv"
)

var asciiTexture = []string{" ", ".", ":", "-", "=", "+", "*", "#", "%", "@"}

func main() {

	if len(os.Args) <= 1 {
		print("Error! Please enter image filename as an argument.")
		return
	}

	filename := os.Args[1]
	im, err := utils.OpenFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	addColors := false
	if len(os.Args) > 2 {
		addColors, err = strconv.ParseBool(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	}

	err = effects.GenerateAsciiFiles(im, asciiTexture, addColors)
	if err != nil {
		log.Fatal(err)
	}
}
