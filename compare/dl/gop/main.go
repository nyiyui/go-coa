package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
}

func download(url string, to string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err2 := Body.Close()
		if err2 != nil {
			err = err2
		}
	}(resp.Body)

	file, err := os.Create(to)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err2 := file.Close()
		if err2 != nil {
			err = err2
		}
	}(file)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	bytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	urls := strings.Split(string(bytes), "\n")
	var wg sync.WaitGroup
	for i, line := range urls {
		split := strings.SplitN(line, " ", 2)
		if line == "" || line[0] == '#' {
			continue
		}
		wg.Add(1)
		go func(i int, url, to string) {
			defer wg.Done()
			err := download(url, to)
			if err != nil {
				log.Printf("error: %s", err)
				return
			}
			log.Printf("downloaded %s to %s", url, to)
		}(i, split[1], split[0])
	}
	wg.Wait()
}
