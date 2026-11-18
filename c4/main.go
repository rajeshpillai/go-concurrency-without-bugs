// c4: a worker pool that can't deadlock. One coordinator owns the map;
// a select either hands out the next job or takes a result, whichever is ready.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"crawlshelf/links"
	"crawlshelf/site"
	"crawlshelf/watch"
)

type result struct {
	url   string
	found []string
	err   error
}

func main() {
	workers := flag.Int("workers", 8, "how many pages to fetch at once")
	trap := flag.Bool("trap", false, "page 137 never answers")
	stuck := flag.Duration("stuck", 30*time.Second, "report what is stuck after this long")
	flag.Parse()
	s := site.Start(*trap)
	defer s.Close()
	done := make(chan struct{})
	go watch.Stuck(*stuck, done)
	peak := watch.Goroutines()
	start := time.Now()

	jobs := make(chan string)
	results := make(chan result)
	for range *workers {
		go func() {
			for url := range jobs {
				found, err := links.Get(http.DefaultClient, s.URL, url)
				results <- result{url, found, err}
			}
		}()
	}

	first := s.URL + "/page/0"
	seen := map[string]bool{first: true}
	todo := []string{first}
	busy := 0
	for len(todo) > 0 || busy > 0 {
		var send chan string // nil: this case is switched off
		var next string
		if len(todo) > 0 {
			send, next = jobs, todo[0]
		}
		select {
		case send <- next:
			todo = todo[1:]
			busy++
		case r := <-results:
			busy--
			for _, f := range r.found {
				if !seen[f] {
					seen[f] = true
					todo = append(todo, f)
				}
			}
		}
	}
	close(jobs)
	close(done)

	fmt.Printf("pages: %d   time: %s   most requests at once: %d   most goroutines: %d\n",
		len(seen), time.Since(start).Round(time.Millisecond), s.Peak.Load(), peak.Stop())
}
