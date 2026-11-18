// A goroutine leak: ask three mirrors for a page, keep the fastest answer, return.
// The two slower goroutines are left trying to send, forever.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// ask pretends to fetch a page from one mirror (mirror i takes i+1 milliseconds).
func ask(i int) string {
	time.Sleep(time.Duration(i+1) * time.Millisecond)
	return fmt.Sprintf("answer from mirror %d", i)
}

// fastest asks three mirrors and returns the first answer.
func fastest(buffered bool) string {
	size := 0
	if buffered {
		size = 3 // room for every answer: nobody has to wait to send
	}
	answers := make(chan string, size)
	for i := range 3 {
		go func() { answers <- ask(i) }()
	}
	return <-answers // take the first; the other two are still sending...
}

func main() {
	buffered := flag.Bool("fixed", false, "use a buffered channel")
	flag.Parse()
	before := runtime.NumGoroutine()
	for range 1000 {
		fastest(*buffered)
	}
	time.Sleep(50 * time.Millisecond) // let the slow mirrors finish
	fmt.Printf("goroutines before: %d   after 1,000 calls: %d\n", before, runtime.NumGoroutine())
	// Go 1.26 experiment: build with GOEXPERIMENT=goroutineleakprofile to get this profile.
	if p := pprof.Lookup("goroutineleak"); p != nil {
		p.WriteTo(os.Stdout, 1) // finds goroutines blocked forever, and where
	}
}
