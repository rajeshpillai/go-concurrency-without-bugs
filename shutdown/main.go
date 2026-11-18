// Graceful shutdown: on Ctrl+C (SIGINT), stop starting new pages, let the ones in flight finish, report.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"crawlshelf/links"
	"crawlshelf/site"
)

func main() {
	s := site.Start(false)
	defer s.Close()
	s.Latency = 200 * time.Millisecond // slower pages, so we can stop in the middle

	// ctx is cancelled when the program gets an interrupt signal.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var done atomic.Int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	for i := range site.Pages {
		select {
		case <-ctx.Done(): // interrupted: start nothing new
		case sem <- struct{}{}:
			wg.Go(func() {
				defer func() { <-sem }()
				url := fmt.Sprintf("%s/page/%d", s.URL, i)
				// in-flight pages get to finish: they don't use the cancelled ctx
				if _, err := links.Get(http.DefaultClient, s.URL, url); err == nil {
					done.Add(1)
				}
			})
			continue
		}
		break
	}
	wg.Wait()
	if ctx.Err() != nil {
		fmt.Printf("interrupted after %d pages: the pages in flight finished, then it stopped cleanly\n", done.Load())
		return
	}
	fmt.Println("all pages done:", done.Load())
}
