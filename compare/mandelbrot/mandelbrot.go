package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"os"
	"runtime/pprof"
	"strconv"
)

func parseArgs() (path string, width, height int64, err error) {
	path = flag.Arg(0)
	width, err = strconv.ParseInt(flag.Arg(1), 10, 32)
	if err != nil {
		return
	}
	height, err = strconv.ParseInt(flag.Arg(2), 10, 32)
	if err != nil {
		return
	}
	return
}

func sq(x complex128) complex128 {
	return x * x
}

func hslToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		return l, l, l
	}

	var v1, v2 float64
	if l < 0.5 {
		v2 = l * (1 + s)
	} else {
		v2 = (l + s) - (s * l)
	}

	v1 = 2*l - v2

	r = hueToRGB(v1, v2, h+(1.0/3.0))
	g = hueToRGB(v1, v2, h)
	b = hueToRGB(v1, v2, h-(1.0/3.0))
	return
}

func hueToRGB(v1, v2, h float64) float64 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	switch {
	case 6*h < 1:
		return v1 + (v2-v1)*6*h
	case 2*h < 1:
		return v2
	case 3*h < 2:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6
	}
	return v1
}

func getColorForIter(i int) color.Color {
	r, g, b := hslToRGB(float64(i)/500, 0.5, 0.5)
	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

func getPixel(x, y float64) int{
	c1 := complex(x, y)
	c2 := complex128(0)
	fmt.Println(c1, c2)
	for n := 1; n < 900; n++ {
		if cmplx.Abs(c2) > 2 {
			return n
		}
		c2 = sq(c2) + c1
	}
	return 0
}

func fillPixel(img *image.RGBA, x, y, _, height int, setWidth float64) {
	fmt.Println(x, y, getPixel(
		(float64(x)-(0.75*setWidth))/(setWidth/4),
		(float64(y)-float64(height)/4-(setWidth/4))/(setWidth/4),
	))
	img.Set(x, y, getColorForIter(getPixel(
		(float64(x)-(0.75*setWidth))/(setWidth/4),
		(float64(y)-float64(height)/4-(setWidth/4))/(setWidth/4),
	)))
}

func getImg(width, height int) (img *image.RGBA) {
	img = image.NewRGBA(image.Rect(0, 0, width, height))
	fillImg(img, width, height)
	return
}

func saveImg(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err2 := file.Close()
		if err2 != nil {
			err = err2
		}
	}(file)
	err = png.Encode(file, img)
	if err != nil {
		return err
	}
	return nil
}

func main_() error {
	path, width, height, err := parseArgs()
	if err != nil {
		return err
	}
	img := getImg(int(width), int(height))
	err = saveImg(img, path)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	err := main_()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s", err)
	}
}
