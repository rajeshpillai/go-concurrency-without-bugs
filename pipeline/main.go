// A pipeline: a stage that hands out page numbers, four fetchers working in parallel (fan-out),
// and one place that collects the results (fan-in). Cancel the context, and every stage stops.
package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"crawlshelf/links"
	"crawlshelf/site"
)

// pages sends the numbers 0..n-1, until ctx is cancelled. It closes its channel when it's done.
func pages(ctx context.Context, n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := range n {
			select {
			case out <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// fetch downloads each page it receives and sends on how many links it had.
func fetch(ctx context.Context, base string, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := range in {
			found, err := links.GetContext(ctx, http.DefaultClient, base, fmt.Sprintf("%s/page/%d", base, i))
			if err != nil {
				continue
			}
			select {
			case out <- len(found):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// merge joins many channels into one (fan-in), and closes it when they are all closed.
func merge(ins ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, in := range ins {
		wg.Go(func() {
			for v := range in {
				out <- v
			}
		})
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

func main() {
	s := site.Start(false)
	defer s.Close()
	before := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	start := time.Now()

	nums := pages(ctx, site.Pages)
	results := merge(fetch(ctx, s.URL, nums), fetch(ctx, s.URL, nums), fetch(ctx, s.URL, nums), fetch(ctx, s.URL, nums))
	got := 0
	for range results {
		got++
		if got == 100 {
			cancel() // we have enough: tell every stage to stop
			break
		}
	}
	for range results { // drain what was already in flight, so every stage can finish
	}
	http.DefaultClient.CloseIdleConnections() // kept-alive connections have goroutines of their own
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("results: %d   time: %s   goroutines before: %d   after: %d\n",
		got, time.Since(start).Round(time.Millisecond), before, runtime.NumGoroutine())
}
