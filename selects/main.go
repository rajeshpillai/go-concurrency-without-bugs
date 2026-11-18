// Two more things select can do: give up after a time, and not wait at all.
package main

import (
	"fmt"
	"time"
)

func main() {
	slow := make(chan string)
	go func() {
		time.Sleep(2 * time.Second)
		slow <- "page"
	}()

	// 1. A timeout: whichever comes first, the page or the clock.
	select {
	case p := <-slow:
		fmt.Println("got", p)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("gave up after 500ms")
	}

	// 2. default: if nothing is ready right now, don't wait at all.
	queue := make(chan string, 1)
	for _, p := range []string{"page 1", "page 2"} {
		select {
		case queue <- p:
			fmt.Println("queued", p)
		default:
			fmt.Println("queue full, dropped", p)
		}
	}
}
