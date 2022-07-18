WIDTH = 20 # for testing
#WIDTH = 4000
HEIGHT = ${WIDTH}

all: mandelbrot_p_go.png mandelbrot_s_go.png;

## Go

mandelbrot_s_go: mandelbrot.go mandelbrot_s.go
	go build -o=mandelbrot_s_go .

mandelbrot_p_go: mandelbrot.go mandelbrot_p.go
	go build -tags=parallel -o=mandelbrot_p_go .

mandelbrot_s_go.png: mandelbrot_s_go
	time ./mandelbrot_s_go mandelbrot_s_go.png ${WIDTH} ${HEIGHT}

mandelbrot_p_go.png: mandelbrot_p_go
	time ./mandelbrot_p_go mandelbrot_p_go.png ${WIDTH} ${HEIGHT}

## Python

mandelbrot_s_py.png: mandelbrot_s.py
	time python3 mandelbrot_s.py mandelbrot_s_py.png ${WIDTH} ${HEIGHT}

## Coa

coa:
	go build -o coa ../../try2/cmd/main.go

mandelbrot_s_coa.png: coa mandelbrot.coa
	time ./coa --parallel=false --path=mandelbrot.coa mandelbrot_s_coa.png ${WIDTH} ${HEIGHT}

mandelbrot_p_coa.png: coa mandelbrot.coa
	time ./coa --parallel=true --path=mandelbrot.coa mandelbrot_p_coa.png ${WIDTH} ${HEIGHT}


.PHONY: all mandelbrot_s_go.png mandelbrot_p_go.png mandelbrot_s_go mandelbrot_p_go mandelbrot_p_coa mandelbrot_s_coa coa mandelbrot_s_py.png
