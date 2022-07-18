package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

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
	for _, line := range strings.Split(string(bytes), "\n") {
		split := strings.SplitN(line, " ", 2)
		if line == "" || line[0] == '#' {
			continue
		}
		to := split[0]
		url := split[1]
		err = download(url, to)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("downloaded %s to %s\n", url, to)
	}
}
