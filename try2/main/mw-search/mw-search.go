package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type result struct {
	name   string
	pageid float64
}

func search(name, baseURL string) (results []result, err error) {
	resp, err := (&http.Client{Timeout: 1 * time.Second}).Get(fmt.Sprintf("https://%s/w/api.php?action=query&list=search&srsearch=%s&format=json", baseURL, name))
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	var rawData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&rawData)
	if err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err2 := Body.Close()
		if err2 != nil {
			err = err2
		}
	}(resp.Body)
	data := rawData["query"].(map[string]interface{})["search"].([]interface{})
	results = make([]result, len(data))
	for i, value := range data {
		results[i] = result{value.(map[string]interface{})["title"].(string), value.(map[string]interface{})["pageid"].(float64)}
	}
	return results, nil
}

func searchAndSave(term, baseURL string) error {
	results, err := search(term, baseURL)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}
	text := fmt.Sprintf("# Search Results for %s (%s)\n\n", term, baseURL)
	for i, result := range results {
		text += fmt.Sprintf("%d. [%s](https://%s/?curid=%d)\n", i+1, result.name, baseURL, result.pageid)
	}
	return os.WriteFile(fmt.Sprintf("search_results_%s.md", term), []byte(text), 0644)
}

func main() {
	start := time.Now()
	err := searchAndSave("concurrency", "ja.wikipedia.org")
	if err != nil {
		panic(err)
	}
	err = searchAndSave("good", "ja.wiktionary.org")
	if err != nil {
		panic(err)
	}
	end := time.Now()
	log.Printf("done in %s", end.Sub(start))
}
