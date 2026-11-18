// checker: check that all 300 pages answer, at most 8 at a time, with errgroup.
// The first error cancels the rest: here, stopping early is what we want.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"crawlshelf/site"
)

func check(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	return nil
}

func main() {
	trap := flag.Bool("trap", false, "page 137 never answers")
	flag.Parse()
	s := site.Start(*trap)
	defer s.Close()
	start := time.Now()

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(8)
	var checked atomic.Int64
	for i := range site.Pages {
		g.Go(func() error {
			if ctx.Err() != nil {
				return ctx.Err() // someone already failed: don't start
			}
			if err := check(ctx, fmt.Sprintf("%s/page/%d", s.URL, i)); err != nil {
				return err
			}
			checked.Add(1)
			return nil
		})
	}
	err := g.Wait()
	fmt.Printf("checked: %d of %d   time: %s   most requests at once: %d\n",
		checked.Load(), site.Pages, time.Since(start).Round(time.Millisecond), s.Peak.Load())
	if err != nil {
		fmt.Println("stopped by:", err)
	}
}
