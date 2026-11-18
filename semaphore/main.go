// A semaphore: a buffered channel with room for 4 tokens. Take one to start, give it back when done.
// At most 4 fetches at a time, with no pool of workers.
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"crawlshelf/links"
	"crawlshelf/site"
)

func main() {
	s := site.Start(false)
	defer s.Close()
	start := time.Now()

	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	for i := range site.Pages {
		sem <- struct{}{} // take a token (waits if all 4 are taken)
		wg.Go(func() {
			defer func() { <-sem }() // give it back
			links.Get(http.DefaultClient, s.URL, fmt.Sprintf("%s/page/%d", s.URL, i))
		})
	}
	wg.Wait()
	fmt.Printf("pages: %d   time: %s   most requests at once: %d\n",
		s.Served.Load(), time.Since(start).Round(time.Millisecond), s.Peak.Load())
}
