package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	start := time.Now()
	result := ""
	for i := int64(0); i < 10001; i++ {
		line := ""
		if i%3 == 0 {
			line += "Fizz"
		}
		if i%5 == 0 {
			line += "Buzz"
		}
		if line == "" {
			line = strconv.FormatInt(i, 10)
		}
		result += line + "\n"
	}
	_, _ = os.Stdout.WriteString(result)
	end := time.Now()
	log.Printf("done in %s", end.Sub(start))
}
