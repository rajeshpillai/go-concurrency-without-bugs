// c0: the simplest crawler. One page at a time.
package main

import (
	"fmt"
	"net/http"
	"time"

	"crawlshelf/links"
	"crawlshelf/site"
)

func main() {
	s := site.Start(false)
	defer s.Close()
	start := time.Now()

	seen := map[string]bool{}
	queue := []string{s.URL + "/page/0"}
	for len(queue) > 0 {
		url := queue[0]
		queue = queue[1:]
		if seen[url] {
			continue
		}
		seen[url] = true
		found, err := links.Get(http.DefaultClient, s.URL, url)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		queue = append(queue, found...)
	}

	fmt.Printf("pages: %d   time: %s   most requests at once: %d\n",
		len(seen), time.Since(start).Round(time.Millisecond), s.Peak.Load())
}
