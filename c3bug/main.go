// c3bug: eight workers and two channels. It looks right. It stops forever.
package main

import (
	"fmt"
	"net/http"
	"time"

	"crawlshelf/links"
	"crawlshelf/site"
	"crawlshelf/watch"
)

func main() {
	s := site.Start(false)
	defer s.Close()
	done := make(chan struct{})
	go watch.Stuck(5*time.Second, done)

	jobs := make(chan string)
	results := make(chan []string)
	for range 8 {
		go func() {
			for url := range jobs {
				found, _ := links.Get(http.DefaultClient, s.URL, url)
				results <- found
			}
		}()
	}

	start := s.URL + "/page/0"
	seen := map[string]bool{start: true}
	jobs <- start
	pending := 1
	for pending > 0 {
		found := <-results
		pending--
		for _, f := range found {
			if !seen[f] {
				seen[f] = true
				pending++
				jobs <- f // waits for a free worker... while every worker waits for us
			}
		}
	}
	close(done)
	fmt.Println("pages:", len(seen))
}
