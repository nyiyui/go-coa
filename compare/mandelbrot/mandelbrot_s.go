//go:build !parallel
// +build !parallel

package main

import (
	"image"
)

func fillImg(img *image.RGBA, width, height int) {
	setWidth := float64(width)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			fillPixel(img, x, y, width, height, setWidth)
			//fmt.Print(x, y, ' ')
		}
	}
}
