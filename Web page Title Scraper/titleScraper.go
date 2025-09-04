package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"strings"
)

// Function to fetch title from a URL
func fetchTitle(url string, wg *sync.WaitGroup, results chan<- string) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		results <- fmt.Sprintf("❌ %s - Error: %v", url, err)
		return
	}
	defer resp.Body.Close()

	// fmt.Println("status code : ", resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)

	// Simple regex to extract <title>...</title>
	re := regexp.MustCompile(`(?s)<title[^>]*>(.*?)</title>`)
	match := re.FindStringSubmatch(string(body))
	// fmt.Println("body : ", string(body))
	// fmt.Println("match : ", match)

	if len(match) > 1 {
		results <- fmt.Sprintf("✅ %s - Title: %s", url, strings.TrimSpace(match[1]))
	} else {
		results <- fmt.Sprintf("⚠️ %s - Title not found - Status Code: %v", url, resp.StatusCode)
	}
}


func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line != "" { 
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}


func main() {
	urlPath := "websites.txt"
	urls, err := readLines(urlPath)
	if !(err == nil) {
		log.Fatalf("error reading urls from %s - %v", urlPath, err)
	}
	// urls := []string{"http://books.toscrape.com/"}

	var wg sync.WaitGroup
	results := make(chan string, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go fetchTitle(url, &wg, results)
	}

	wg.Wait()
	close(results)

	for result := range results {
		fmt.Println(result)
	}
}
