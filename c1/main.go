// c1: a goroutine for every link. Fast, and broken: the map is shared with no lock.
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

	seen := map[string]bool{}
	var wg sync.WaitGroup
	var crawl func(url string)
	crawl = func(url string) {
		if seen[url] {
			return
		}
		seen[url] = true
		found, err := links.Get(http.DefaultClient, s.URL, url)
		if err != nil {
			return
		}
		for _, f := range found {
			wg.Go(func() { crawl(f) })
		}
	}
	wg.Go(func() { crawl(s.URL + "/page/0") })
	wg.Wait()

	fmt.Printf("pages: %d   fetched: %d   time: %s   most requests at once: %d\n",
		len(seen), s.Served.Load(), time.Since(start).Round(time.Millisecond), s.Peak.Load())
}
