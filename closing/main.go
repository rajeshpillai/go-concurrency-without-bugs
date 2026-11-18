// What closing a channel does: range stops, receives get the zero value, and sends panic.
package main

import (
	"fmt"
	"os"
)

func main() {
	pages := make(chan string, 3)
	pages <- "page 1"
	pages <- "page 2"
	close(pages) // no more values will come

	for p := range pages { // range keeps receiving until the channel is closed and empty
		fmt.Println("got", p)
	}
	p, ok := <-pages // closed and empty: the zero value, and ok is false
	fmt.Printf("after close: %q, ok=%v\n", p, ok)

	if len(os.Args) > 1 && os.Args[1] == "send" {
		pages <- "page 3" // sending on a closed channel panics
	}
}
