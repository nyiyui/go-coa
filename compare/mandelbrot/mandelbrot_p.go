//go:build parallel
// +build parallel

package main

import (
	"flag"
	"image"
	"sync"
)

var _blockSize = flag.Int("block", 1000, "size of block used to partition dat")

func _fillImgBlock(img *image.RGBA, wg *sync.WaitGroup, widthPerBlock, heightPerBlock, width, height, xS, yS int, setWidth float64) {
	// xS = X of Starting point
	defer wg.Done()
	for x := xS; x < xS+widthPerBlock; x++ {
		for y := yS; y < yS+heightPerBlock; y++ {
			fillPixel(img, x, y, width, height, setWidth)
		}
	}
}

func fillImg(img *image.RGBA, width, height int) {
	setWidth := float64(width)
	blockSize := *_blockSize
	var wg sync.WaitGroup
	wg.Add((width / blockSize) * (height / blockSize))
	for xB := 0; xB < width/blockSize; xB++ {
		for yB := 0; yB < height/blockSize; yB++ {
			go _fillImgBlock(img, &wg, blockSize, blockSize, width, height, xB*blockSize, yB*blockSize, setWidth)
		}
	}
	wg.Wait()
}
