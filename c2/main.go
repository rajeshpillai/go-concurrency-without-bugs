// c2: a goroutine for every link, and a mutex around the map. Correct, and very fast.
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"crawlshelf/links"
	"crawlshelf/site"
	"crawlshelf/watch"
)

func main() {
	s := site.Start(false)
	defer s.Close()
	peak := watch.Goroutines()
	start := time.Now()

	var mu sync.Mutex
	seen := map[string]bool{}
	// firstVisit marks url as seen, and reports whether this is the first time.
	firstVisit := func(url string) bool {
		mu.Lock()
		defer mu.Unlock()
		if seen[url] {
			return false
		}
		seen[url] = true
		return true
	}

	var wg sync.WaitGroup
	var crawl func(url string)
	crawl = func(url string) {
		if !firstVisit(url) {
			return
		}
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

	fmt.Printf("pages: %d   time: %s   most requests at once: %d   most goroutines: %d\n",
		len(seen), time.Since(start).Round(time.Millisecond), s.Peak.Load(), peak.Stop())
}
